package main

import (
	"fmt"
	"net/http"

	"chitchat/webserver"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	fs "github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html"
)

// const ws = websocket

type TypeJSON = map[string]interface{}
type Ctx = *fiber.Ctx

func main() {
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
	app.Get("/app", func(c Ctx) error {
		return c.Render("app", nil)
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
