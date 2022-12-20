package mainapp

import (
	"chitchat/webserver"
	"fmt"

	"github.com/fasthttp/websocket"
)

type UserSessionState struct {
	MyUsername  string
	ChatSession map[string]TypeJSON
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
	state := UserSessionState{
		ChatSession: map[string]TypeJSON{},
	}

	defer func() {
		for _, session := range state.ChatSession {
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
				fmt.Println("print ci user_sessions", ci)
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
				session, isOk := state.ChatSession[username]
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
						state.ChatSession[username] = session
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
				session, isOk := state.ChatSession[ci.Payload["username"].(string)]
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
				session, isOk := state.ChatSession[ci.Payload["username"].(string)]
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
				state.ChatSession[pl["username"].(string)] = TypeJSON{
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
