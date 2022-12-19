package mainapp

func (c *ClientOutput) toEstablishChatSession() EstablishChatSession {
	return EstablishChatSession{
		Username: c.getString(&c.Payload, "username"),
		G:        c.getInt(&c.Payload, "g"),
		N:        c.getInt(&c.Payload, "n"),
	}
}
