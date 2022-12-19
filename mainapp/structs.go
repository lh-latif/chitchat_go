package mainapp

import (
	"math/rand"
	"time"
)

type TypeJSON = map[string]interface{}

type Messaging struct {
	// Command string   `json:"command"`
	// Payload TypeJSON `json:"payload"`
}

type WsWriterChan = chan ClientOutput

type UserChatChannel = chan UserChatMsg

type ClientOutput struct {
	Messaging
	Command string   `json:"command"`
	Payload TypeJSON `json:"payload"`
}

type ClientInput struct {
	Messaging
	Command string   `json:"command"`
	Payload TypeJSON `json:"payload"`
}

type UserChatMsg struct {
	Messaging
	Command string
	Payload TypeJSON
}

type UserWsState struct {
	MyUsername string
	Session    map[string]TypeJSON
}

type ClientInputPayload struct{ TypeJSON }

func NumberRandom() uint8 {
	return uint8(rand.New(rand.NewSource(65)).Int())
}

type UserConnectionState struct {
	Status   string
	Username string
	G        int
	N        uint8
}

type RegisterPayload struct {
	Username string `json:"username"`
}

type SendChatPayload struct {
	Username string    `json:"username"`
	Time     time.Time `json:"time"`
	Content  string    `json:"content"`
}

type SharedKeyPayload struct {
	G int
	N uint8
}

type InputEstablishChatSession struct {
	Username string
}

type EstablishChatSession struct {
	Username string
	G        int
	N        int
}

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
