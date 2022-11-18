package webserver

type SessionProcMsg struct {
	Name    string
	Payload interface{}
}

type SessionState struct {
	ListUser ListUserMap
}

type UserSessionChan = chan SessionProcMsg

type ListUserMap = map[string]UserSessionChan

type ChatProto struct {
	Session string
}

func SessionProcessStart() (chan SessionProcMsg, bool) {
	return session_start()
}

type TypeJSON = map[string]interface{}

func session_start() (chan SessionProcMsg, bool) {
	feedback_chan := make(chan chan SessionProcMsg)
	feedback_err := make(chan int)

	go func() {
		input := make(chan SessionProcMsg)
		feedback_chan <- input
		var state SessionState = init_state(feedback_err)
		for {
			msg := <-input
			switch msg.Name {
			case "register":
				p := msg.Payload.(TypeJSON)
				name := p["name"].(string)
				session_chan := p["session_chan"].(UserSessionChan)
				from := p["result"].(chan int)
				state.ListUser[name] = session_chan
				from <- 1
				continue
			case "unregister":
				continue
			case "get_user":
				username := msg.Payload.(TypeJSON)["username"].(string)
				from := msg.Payload.(TypeJSON)["result"].(chan UserSessionChan)
				user, isOk := state.ListUser[username]
				if isOk {
					from <- user
				} else {
					from <- nil
				}
				continue
			case "list_user":
				var users []string
				for key, _ := range state.ListUser {
					users = append(users, key)
				}
				from := msg.Payload.(TypeJSON)["result"].(chan []string)
				from <- users
				continue
			}

		}
	}()
	var is_error bool
	var channel chan SessionProcMsg
	select {
	case channel = <-feedback_chan:
		is_error = false
	case <-feedback_err:
		is_error = true
	}
	close(feedback_chan)
	close(feedback_err)
	if is_error {
		return nil, true
	} else {
		return channel, false
	}
}

func init_state(feedback_err chan int) SessionState {
	// defer (feedback_err <- 1);
	_ = feedback_err
	var state SessionState = SessionState{
		ListUser: ListUserMap{},
	}
	return state
}

func send_chat() {

}
