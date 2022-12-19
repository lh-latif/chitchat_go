package mainapp

import "github.com/fasthttp/websocket"

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
