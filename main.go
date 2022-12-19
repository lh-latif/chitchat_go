package main

import (
	"net/http"

	db "chitchat/database"
	"chitchat/mainapp"
	"chitchat/webserver"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	fs "github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/argon2"
)

type ErrorText string

type UserRegisterBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserSigninBody struct {
	UserRegisterBody
}

type UserContactBody struct {
	Id          uint   `json:"contact_id"`
	ContactName string `json:"contact_name"`
}

type SearchUserBody struct {
	Username string `json:"username"`
}

func (e ErrorText) Error() string {
	return string(e)
}

// const ws = websocket

type TypeJSON = map[string]interface{}
type Ctx = *fiber.Ctx

func main() {
	dbConn := db.ConnectDB()
	_ = dbConn
	sessionProcess, isError := webserver.SessionProcessStart()
	_ = isError
	_ = sessionProcess
	templateEngine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{
		Views: templateEngine,
	})

	wsUpgrader := websocket.FastHTTPUpgrader{
		ReadBufferSize:  1023000,
		WriteBufferSize: 1023000,
	}

	app.Get("/debug/users_session", func(c Ctx) error {
		response := make(chan []string)
		msg := webserver.SessionProcMsg{
			Name: "list_user",
			Payload: TypeJSON{
				"result": response,
			},
		}
		sessionProcess <- msg
		result := <-response
		close(response)
		return c.JSON(TypeJSON{"data": result})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World!")
	})

	app.Use("/assets", fs.New(fs.Config{
		Root: http.Dir("./web_assets"),
	}))

	app.Get("/app*", func(c Ctx) error {
		return c.Render("app", nil)
	})

	app.Post("/user/register", func(c Ctx) error {
		body := UserRegisterBody{}
		err := c.BodyParser(&body)
		_ = err
		newUser := db.User{
			Username: body.Username,
			Hash:     string(argon2.IDKey([]byte(body.Password), []byte("programming in golang"), 10, 64*1024, 4, 32)),
		}
		dbConn.Create(&newUser)
		return err
	})

	app.Get("/user/index", func(c Ctx) error {
		var users []db.User
		dbConn.Find(&users)
		return c.JSON(TypeJSON{
			"users": users,
		})
	})

	app.Post("/user/signin", func(c Ctx) error {
		body := UserSigninBody{}
		c.BodyParser(&body)
		var user db.User
		dbConn.Find(&user, "username = ?", body.Username)
		if user.Username != "" && userPass(&body, &user) {
			token, err := webserver.InitiateToken(jwt.MapClaims{
				"username": user.Username,
			})
			if err != nil {
				return ErrorText("token signing error")
			}
			return c.JSON(TypeJSON{
				"token": token,
			})
		} else if user.Username != "" {
			return ErrorText("Password mismatch")
		} else {
			return ErrorText("Username not found")
		}
	})

	app.Get("/ws", func(c *fiber.Ctx) error {
		err := wsUpgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
			defer conn.Close()

			resultChan := make(chan mainapp.WsWriterChan)
			go mainapp.WsWriter(conn, resultChan)
			writerChan := <-resultChan

			wsChan := make(chan (chan mainapp.ClientInput))
			go mainapp.UserSessionProcess(conn, wsChan, sessionProcess, writerChan)
			inSessChan := <-wsChan
			if inSessChan == nil {
				return
			}
			defer func() {
				// from <- "reader_stopped"
				inSessChan <- mainapp.ClientInput{
					Command: "client_disconnect",
				}
				writerChan <- mainapp.ClientOutput{
					Command: "client_disconnect",
				}
			}()

			var input mainapp.ClientInput
			for {

				err := conn.ReadJSON(&input)
				if err != nil {
					// fmt.Println(err.Error())
					break
				}
				// fmt.Println(input)
				// command := input["command"].(string)
				switch input.Command {
				case "register":
					inSessChan <- input
				case "println":
					// fmt.Println(input.Payload)
					continue
				case "send_chat":
					inSessChan <- input
				case "establish_chat_session":
					inSessChan <- input
				case "send_public_key":
					inSessChan <- input
				case "send_ready":
					inSessChan <- input
				}
			}
		})
		return err
	})

	jwtGroup := app.Group("/", webserver.JwtCheckMiddleware(dbConn))

	jwtGroup.Get("/user/me", func(c Ctx) error {
		user := c.Context().UserValue("user").(db.User)
		return c.JSON(TypeJSON{
			"data": user,
		})
	})

	jwtGroup.Post("/search_user", func(c Ctx) error {
		body := SearchUserBody{}
		c.BodyParser(&body)
		users := []db.User{}
		dbConn.Where("username LIKE ?", "%"+body.Username+"%").Find(&users)
		return c.JSON(TypeJSON{
			"data": users,
		})
	})

	jwtGroup.Post("/user/contact", func(c Ctx) error {
		user := c.Context().UserValue("user").(db.User)
		body := UserContactBody{}
		c.BodyParser(&body)
		contact := db.UserContact{
			UserId:      user.Id,
			ContactId:   body.Id,
			ContactName: body.ContactName,
		}
		dbConn.Create(&contact)
		return c.SendString("ok")
	})

	jwtGroup.Get("/user/contact", func(c Ctx) error {
		user := c.Context().UserValue("user").(db.User)
		contacts := []db.UserContact{}
		dbConn.Where("user_id = ?", user.Id).Find(&contacts)
		listUsers := []db.User{}
		userId := []uint{}
		for _, contact := range contacts {
			userId = append(userId, contact.ContactId)
		}

		dbConn.Find(&listUsers, userId)
		// fmt.Println(userId, &listUsers)
		var response = []TypeJSON{}
		for _, contact := range contacts {
			var user db.User
			isNil := true
			for _, userx := range listUsers {
				// fmt.Println(userx.Id, contact.ContactId)
				if userx.Id == contact.ContactId {
					user = userx
					isNil = false
					break
				} else {
					continue
				}
			}
			if isNil {
				return c.Status(400).SendString("Error")
			}
			// fmt.Println(user)
			response = append(response, TypeJSON{
				"id":           contact.Id,
				"contact_name": contact.ContactName,
				"username":     user.Username,
			})
		}
		return c.JSON(TypeJSON{
			"data": response,
		})
	})

	app.Listen(":8000")
}

func userPass(body *UserSigninBody, user *db.User) bool {
	return string(argon2.IDKey([]byte(body.Password), []byte("programming in golang"), 10, 64*1024, 4, 32)) == user.Hash
}
