package message

import (
	"time"
)

type Message struct {
	ID           string
	AckID        string
	Action       string
	Data         []byte
	Attributes   map[string]string
	PublishTime  time.Time
	Deadline     time.Time
	Duration     time.Duration
	MessageState State
	Attempts     int
	Sender       string
}

func NewMessage(id, ack_id, action, sender string, data []byte, attributes map[string]string, deadline time.Duration) *Message {
	now := time.Now()
	return &Message{
		ID:           id,
		AckID:        ack_id,
		Action:       action,
		Data:         data,
		Attributes:   attributes,
		PublishTime:  now,
		Deadline:     now.Add(deadline),
		Duration:     deadline,
		MessageState: PUBLISH,
		Sender:       sender,
		Attempts:     0,
	}
}

type State int

const (
	PUBLISH State = iota
	OUTGOING
	SENT
	ACK
	NACK
)
