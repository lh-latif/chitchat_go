package mainapp

import (
	"chitchat/webserver"
	"fmt"
)

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
