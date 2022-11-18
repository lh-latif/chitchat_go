package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

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

			resultChan := make(chan WsWriterChan)
			go WsWriter(conn, resultChan)
			writerChan := <-resultChan

			wsChan := make(chan (chan ClientInput))
			go UserSessionProcess(conn, wsChan, sessionProcess, writerChan)
			inSessChan := <-wsChan
			if inSessChan == nil {
				return
			}
			defer func() {
				// from <- "reader_stopped"
				inSessChan <- ClientInput{
					Command: "client_disconnect",
				}
				writerChan <- ClientOutput{
					Command: "client_disconnect",
				}
			}()

			var input ClientInput
			for {

				err := conn.ReadJSON(&input)
				if err != nil {
					fmt.Println(err.Error())
					break
				}
				fmt.Println(input.RegisterPayload())
				// command := input["command"].(string)
				switch input.Command {
				case "register":
					inSessChan <- input
				case "println":
					fmt.Println(input.Payload)
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
		fmt.Println(userId, &listUsers)
		var response = []TypeJSON{}
		for _, contact := range contacts {
			var user db.User
			isNil := true
			for _, userx := range listUsers {
				fmt.Println(userx.Id, contact.ContactId)
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
			fmt.Println(user)
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

type UserWsState struct {
	MyUsername string
	Session    map[string]TypeJSON
}

type ClientInput struct {
	Command string   `json:"command"`
	Payload TypeJSON `json:"payload"`
}

type ClientInputPayload struct{ TypeJSON }

func (p *ClientInput) getValue(key string) interface{} {
	return p.Payload[key]
}

func (p *ClientInput) getString(key string) string {
	return p.getValue(key).(string)
}

func (p *ClientInput) getTime(key string) time.Time {
	v, _ := time.Parse(time.RFC3339, p.getString(key))
	return v
}

func (p *ClientInput) RegisterPayload() RegisterPayload {
	return RegisterPayload{
		Username: p.getString("username"),
	}
}

func (p *ClientInput) SendChatPayload() SendChatPayload {
	return SendChatPayload{
		Username: p.getString("username"),
		Time:     p.getTime("time"),
		Content:  p.getString("content"),
	}
}

type RegisterPayload struct {
	Username string `json:"username"`
}

type SendChatPayload struct {
	Username string    `json:"username"`
	Time     time.Time `json:"time"`
	Content  string    `json:"content"`
}

func UserSessionProcess(
	conn *websocket.Conn,
	from chan (chan ClientInput),
	sessionProcess chan webserver.SessionProcMsg,
	wsWriterChan WsWriterChan,
) {
	inWsChan := make(chan ClientInput)
	from <- inWsChan
	inSessionChan := make(webserver.UserSessionChan)
	state := UserWsState{
		Session: map[string]TypeJSON{},
	}

	defer func() {
		for _, session := range state.Session {
			chatChannel := session["chat_session"].(UserChatChannel)
			chatChannel <- UserChatMsg{
				Command: "client_disconnect",
			}
		}
	}()

	for {
		select {
		case ci := <-inWsChan:
			switch ci.Command {
			case "client_disconnect":
				break
			case "register":
				payload := ci.RegisterPayload()

				response := make(chan int)
				sessionProcess <- webserver.SessionProcMsg{
					Name: "register",
					Payload: TypeJSON{
						"name":         payload.Username,
						"result":       response,
						"session_chan": inSessionChan,
					},
				}
				result := <-response
				close(response)
				state.MyUsername = payload.Username
				if result == 1 {
					fmt.Println("ok")
				}
			case "establish_chat_session":
				// payload := ci.SendChatPayload()
				payload := ci.Payload
				username := payload["username"].(string)
				// _ = payload
				session, isOk := state.Session[username]
				if !isOk {
					// buat sessionnya
					response := make(chan webserver.UserSessionChan)
					sessionProcess <- webserver.SessionProcMsg{
						Name: "get_user",
						Payload: TypeJSON{
							"username": username,
							"result":   response,
						},
					}
					result := <-response
					close(response)
					if result != nil {
						response := make(chan UserChatChannel)
						go UserChatSession(
							true,
							response,
							wsWriterChan,
							state.MyUsername,
							username,
							result,
							nil,
						)
						userChatSessionChan := <-response
						close(response)
						session = TypeJSON{
							"chat_session": userChatSessionChan,
						}
						state.Session[username] = session
					} else {
						fmt.Println(result, "error result == nil")
						continue
					}
				} else {
					fmt.Println("found")
				}
				// _ = session
				// chatSession := session["chat_session"].(UserChatChannel)
				// chatSession <- UserChatMsg{
				// 	Command: "send_chat",
				// }
			case "send_public_key":
				session, isOk := state.Session[ci.Payload["username"].(string)]
				if isOk {
					session["chat_session"].(UserChatChannel) <- UserChatMsg{
						Command: "send_public_key",
						Payload: TypeJSON{
							"public_key": ci.Payload["public_key"].(string),
						},
					}
				} else {
					fmt.Println("send_public_key error, ", ci, state)
				}

			case "send_chat":
				session, isOk := state.Session[ci.Payload["username"].(string)]
				if isOk {
					session["chat_session"].(UserChatChannel) <- UserChatMsg{
						Command: "send_chat",
						Payload: ci.Payload,
					}
				}
			}

			continue
		case si := <-inSessionChan:
			_ = si
			switch si.Name {
			case "establish":
				pl := si.Payload.(TypeJSON)

				response := make(chan UserChatChannel)
				targetInputChan := pl["input_chan"].(UserChatChannel)
				go UserChatSession(
					false,
					response,
					wsWriterChan,
					state.MyUsername,
					pl["username"].(string),
					nil,
					targetInputChan,
				)
				userChatSessionChan := <-response
				close(response)

				from := pl["result"].(chan UserChatChannel)
				from <- userChatSessionChan
				state.Session[pl["username"].(string)] = TypeJSON{
					"chat_session": userChatSessionChan,
				}
			}
			// start session
			// end_session
			// want to send
			// want to confirm
			continue
		}
	}

}

func WsWriter(conn *websocket.Conn, from chan WsWriterChan) {
	inChan := make(WsWriterChan)
	from <- inChan
	for {
		in := <-inChan
		switch in.Command {
		case "establish_chat_session":
			conn.WriteJSON(TypeJSON{
				"command": "establish_chat_session",
				"payload": TypeJSON{
					"sender": in.Payload["username"],
					"g":      in.Payload["g"],
					"n":      in.Payload["n"],
				},
			})
		case "write_received_chat":
			conn.WriteJSON(TypeJSON{
				"command": "receive_chat",
				"payload": TypeJSON{
					"sender":  in.Payload["username"],
					"content": in.Payload["content"],
				},
			})
		case "write_public_key":
			conn.WriteJSON(TypeJSON{
				"command": "receive_public_key",
				"payload": TypeJSON{
					"sender":     in.Payload["username"],
					"public_key": in.Payload["public_key"],
				},
			})
		case "write_ready":
			conn.WriteJSON(TypeJSON{
				"command": "receive_ready",
				"payload": in.Payload,
			})
		}
	}
}

type UserChatChannel = chan UserChatMsg

type UserChatMsg struct {
	Command string
	Payload TypeJSON
}

func (u *UserChatMsg) SharedKeyPayload() SharedKeyPayload {
	return SharedKeyPayload{
		G: u.Payload["g"].(int),
		N: u.Payload["n"].(uint8),
	}
}

type SharedKeyPayload struct {
	G int
	N uint8
}

func UserChatSession(
	shouldEstablish bool,
	from chan UserChatChannel,
	wsWriterChan WsWriterChan,
	myUsername string,
	username string,
	targetSession webserver.UserSessionChan,
	targetChanSession UserChatChannel,
) {
	fmt.Println("chat session initiated")
	// generator g = prime number
	// number n = 500byte number (should be)
	inputChan := make(UserChatChannel)
	from <- inputChan
	state := UserConnectionState{
		Username: username,
	}
	// send signal to recipient chat session
	if shouldEstablish {

		response := make(chan UserChatChannel)
		targetSession <- webserver.SessionProcMsg{
			Name: "establish",
			Payload: TypeJSON{
				"username":   myUsername,
				"result":     response,
				"input_chan": inputChan,
			},
		}
		result := <-response
		close(response)

		if result != nil {
			targetChanSession = result
		}

		// after shouldEstablish then send them g and n
		state.G = 7
		state.N = NumberRandom()
		targetChanSession <- UserChatMsg{
			Command: "share_public_key",
			Payload: TypeJSON{
				"g": state.G,
				"n": state.N,
			},
		}
		wsWriterChan <- ClientOutput{
			Command: "establish_chat_session",
			Payload: TypeJSON{
				"g":        state.G,
				"n":        state.N,
				"username": username,
			},
		}
		state.Status = "ready"
	}

	defer func() {
		close(inputChan)
	}()

	for {
		input := <-inputChan
		switch input.Command {
		case "client_disconnect":
			targetChanSession <- UserChatMsg{
				Command: "break_off_sayonara",
			}
			break
		case "break_off_sayonara":
			// should tell user-session if
			// chat session is over
			break
		case "share_public_key":
			payload := input.SharedKeyPayload()
			state.G = payload.G
			state.N = payload.N
			state.Status = "ready"
			wsWriterChan <- ClientOutput{
				Command: "establish_chat_session",
				Payload: TypeJSON{
					"g":        state.G,
					"n":        state.N,
					"username": username,
				},
			}
		case "send_chat":
			targetChanSession <- UserChatMsg{
				Command: "forward_chat_to_client",
				Payload: input.Payload,
			}
		case "forward_chat_to_client":
			wsWriterChan <- ClientOutput{
				Command: "write_received_chat",
				Payload: TypeJSON{
					"content":  input.Payload["content"],
					"username": username,
				},
			}
		case "send_public_key":
			fmt.Println(input)
			targetChanSession <- UserChatMsg{
				Command: "forward_public_key",
				Payload: input.Payload,
			}
		case "forward_public_key":
			wsWriterChan <- ClientOutput{
				Command: "write_public_key",
				Payload: TypeJSON{
					"public_key": input.Payload["public_key"],
					"username":   username,
				},
			}
			continue
		}
	}

}

func NumberRandom() uint8 {
	return uint8(rand.New(rand.NewSource(65)).Int())
}

type UserConnectionState struct {
	Status   string
	Username string
	G        int
	N        uint8
}

type WsWriterChan = chan ClientOutput

type ClientOutput struct {
	Command string
	Payload TypeJSON
}
