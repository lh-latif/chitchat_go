package main

import (
	"fmt"
	"net/http"

	db "chitchat/database"
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
	Username string `json="username"`
	Password string `json="password"`
}

type UserSigninBody struct {
	UserRegisterBody
}

type UserContactBody struct {
	Id          uint   `json="id"`
	ContactName string `json="contact_name"`
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
	userGroup := app.Group("/user", webserver.JwtCheckMiddleware(dbConn))

	userGroup.Post("/contact", func(c Ctx) error {
		body := UserContactBody{}
		return c.BodyParser(&body)
	})

	userGroup.Get("/contact", func(c Ctx) error {
		user := c.Context().UserValue("user").(db.User)
		contacts := []db.UserContact{}
		dbConn.Where("user_id = ?", user.Id).Find(&contacts)
		return c.JSON(TypeJSON{
			"data": contacts,
		})
	})

	app.Get("/ws", func(c *fiber.Ctx) error {
		err := wsUpgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
			defer conn.Close()
			var msg TypeJSON = TypeJSON{}
			for {
				err := conn.ReadJSON(&msg)
				if err != nil {
					break
				}
				switch msg["command"] {
				case "register":
					name := msg["payload"].(TypeJSON)["name"].(string)
					response := make(chan int)
					session_chan := make(webserver.ChannelChatProto)
					sessionProcess <- webserver.SessionProcMsg{
						Name: "register",
						Payload: TypeJSON{
							"name":         name,
							"result":       response,
							"session_chan": session_chan,
						},
					}
					result := <-response
					if result == 1 {
						fmt.Println("ok")
					}
				}
				fmt.Println(msg)
			}

		})
		return err
	})

	app.Listen(":8000")
}

func userPass(body *UserSigninBody, user *db.User) bool {
	return string(argon2.IDKey([]byte(body.Password), []byte("programming in golang"), 10, 64*1024, 4, 32)) == user.Hash
}
