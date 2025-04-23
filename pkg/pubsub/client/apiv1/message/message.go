package message

import (
	"time"
)

type Message struct {
	ID            string
	AckID         string
	Action        string
	Data          []byte
	Attributes    map[string]string
	PublishedTime time.Time
	AckChan       chan string
}

func NewMessage(action string, data []byte) *Message {
	return &Message{
		Action: action,
		Data:   data,
	}
}

func (m *Message) Ack() {
	m.AckChan <- m.AckID
}

func (m *Message) Nack() {}
