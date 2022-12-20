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
