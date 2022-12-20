package mainapp

import "time"

func (u *UserChatMsg) SharedKeyPayload() SharedKeyPayload {
	return SharedKeyPayload{
		G: u.Payload["g"].(int),
		N: u.Payload["n"].(uint8),
	}
}

func (p *Messaging) getValue(src *TypeJSON, key string) interface{} {
	var srcx TypeJSON = *src
	return srcx[key]
}

func (p *Messaging) getString(src *TypeJSON, key string) string {
	return p.getValue(src, key).(string)
}

func (p *Messaging) getTime(src *TypeJSON, key string) time.Time {
	v, _ := time.Parse(time.RFC3339, p.getString(src, key))
	return v
}

func (p *Messaging) getInt(src *TypeJSON, key string) int {
	return p.getValue(src, key).(int)
}

func (p *ClientInput) RegisterPayload() RegisterPayload {
	// fmt.Println(p)
	return RegisterPayload{
		Username: p.getString(&p.Payload, "username"),
	}
}

func (p *ClientInput) SendChatPayload() SendChatPayload {
	return SendChatPayload{
		Username: p.getString(&p.Payload, "username"),
		Time:     p.getTime(&p.Payload, "time"),
		Content:  p.getString(&p.Payload, "content"),
	}
}
