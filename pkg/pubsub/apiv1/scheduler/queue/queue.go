package queue

import (
	"container/list"
	"sync"
	"time"

	"github.com/axelburling/dfs/pkg/pubsub/apiv1/message"
)

const MAX_ATTEMPTS = 10

type Queue struct {
	mu     sync.Mutex
	queue  *list.List
	sent   map[string]message.Message
	Active bool
}

func NewQueue() *Queue {
	return &Queue{queue: list.New(), Active: false, sent: make(map[string]message.Message)}
}

func (q *Queue) AddMessage(msg *message.Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue.PushBack(msg)
}

func (q *Queue) AddMessages(msgs []*message.Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, msg := range msgs {
		q.queue.PushBack(msg)
	}
}

func (q *Queue) GetMessages(n int) []*message.Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.Active = true

	messages := []*message.Message{}
	count := 0

	for e := q.queue.Front(); e != nil && count < n; e = e.Next() {
		msg := e.Value.(*message.Message)

		msg.MessageState = message.SENT
		messages = append(messages, msg)
		count++
		q.queue.Remove(e)
		q.sent[msg.AckID] = *msg
	}

	return messages
}

func (q *Queue) AckMessages(ids []string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, id := range ids {
		if _, ok := q.sent[id]; ok {
			delete(q.sent, id)
		}
	}

	// for e := q.queue.Front(); e != nil; {
	// 	msg := e.Value.(*message.Message)
	// 	if idSet[msg.AckID] {
	// 		next := e.Next()
	// 		q.queue.Remove(e)
	// 		e = next
	// 	} else {
	// 		e = e.Next()
	// 	}
	// }
}

func (q *Queue) NackMessages(ids []string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	for e := q.queue.Front(); e != nil; {
		msg := e.Value.(*message.Message)

		// Re-queue the message if it has expired and was not acknowledged.
		if idSet[msg.AckID] {
			q.queue.Remove(e)
			msg.Deadline = now.Add(msg.Duration) // Reset the acknowledgment deadline.
			q.queue.PushFront(msg)
		} else {
			e = e.Next()
		}
	}
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue.Init()
}

func (q *Queue) RequeueExpiredMessages() {
	q.mu.Lock()
	now := time.Now()

	var toRequeue []*message.Message

	// Collect expired messages first (without modifying the map)
	for id, msg := range q.sent {
		if msg.Attempts >= MAX_ATTEMPTS {
			delete(q.sent, id) // Safe deletion since we are only removing
			continue
		}

		if msg.Deadline.Before(now) {
			msg.Deadline = now.Add(msg.Duration)
			msg.MessageState = message.PUBLISH
			msg.Attempts++
			toRequeue = append(toRequeue, &msg)
			delete(q.sent, id) // Move message out of `sent`
		}
	}

	q.mu.Unlock() // Unlock before adding messages to queue

	// Add requeued messages outside critical section
	if len(toRequeue) > 0 {
		q.mu.Lock()
		for _, msg := range toRequeue {
			q.queue.PushBack(msg)
		}
		q.mu.Unlock()
	}
}
