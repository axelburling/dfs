package scheduler

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/axelburling/dfs/pkg/pubsub/apiv1/message"
	"github.com/axelburling/dfs/pkg/pubsub/apiv1/scheduler/queue"
)

var (
	ErrTopicNotExist        = errors.New("Topic does not exist")
	ErrSubscriptionNotExist = errors.New("Subscription does not exist")
)

type Scheduler struct {
	topics         map[string]*topic
	mu             sync.Mutex
	commandChannel chan interface{}
	interval       time.Duration
}

type topic struct {
	mu          sync.Mutex
	subscribers map[string]*queue.Queue
}

type createTopicCommand struct {
	TopicName string
	Response  chan *topic
}

type addSubscriberCommand struct {
	TopicName      string
	SubscriberName string
	Response       chan *queue.Queue
}

type publishMessageCommand struct {
	TopicName string
	Messages  []*message.Message
	Response  chan error
}

type getMessagesCommand struct {
	TopicName      string
	SubscriberName string
	Count          int
	Response       chan []*message.Message
}

type ackMessagesCommand struct {
	TopicName      string
	SubscriberName string
	AckIDs         []string
	Response       chan error
}

type getSubscriberQueueCommand struct {
	TopicName      string
	SubscriberName string
	Response       chan *queue.Queue
}

type getTopicCommand struct {
	TopicName string
	Response  chan *topic
}

type removeSubscriberCommand struct {
	TopicName      string
	SubscriberName string
	Response       chan error
}

type deleteTopicCommand struct {
	TopicName string
	Response  chan error
}

func NewScheduler(interval time.Duration) *Scheduler {
	return &Scheduler{
		topics:         make(map[string]*topic),
		commandChannel: make(chan interface{}),
		interval:       interval,
	}
}

func (s *Scheduler) Start(stop chan os.Signal) {
	go func() {
		for {
			select {
			case cmd, ok := <-s.commandChannel:
				if !ok {
					continue
				}
				switch c := cmd.(type) {
				case *createTopicCommand:
					s.handleCreateTopic(c)
				case *addSubscriberCommand:
					s.handleAddSubscriber(c)
				case *publishMessageCommand:
					s.handlePublishMessage(c)
				case *getMessagesCommand:
					s.handleGetMessages(c)
				case *ackMessagesCommand:
					s.handleAckMessages(c)
				case *getSubscriberQueueCommand:
					s.handleGetSubscriberQueue(c)
				case *getTopicCommand:
					s.handleGetTopic(c)
				case *removeSubscriberCommand:
					s.handleRemoveSubscriber(c)
				case *deleteTopicCommand:
					s.handleDeleteTopic(c)
				}
			case <-stop:
				fmt.Println("stopping scheduler")
				break
			}
		}
	}()

	s.requeueMonitor(s.interval, stop)
}

func (s *Scheduler) Empty(topicName string) bool {
	topic, ok := s.topics[topicName]

	if !ok {
		return true
	}

	if len(topic.subscribers) == 0 {
		return true
	}

	return false
}

func (s *Scheduler) handleGetTopic(cmd *getTopicCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, exists := s.topics[cmd.TopicName]
	if !exists {
		cmd.Response <- nil
		return
	}

	cmd.Response <- topic
}

func (s *Scheduler) handleAckMessages(cmd *ackMessagesCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, exists := s.topics[cmd.TopicName]
	if !exists {
		cmd.Response <- ErrTopicNotExist
		return
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	q, exists := topic.subscribers[cmd.SubscriberName]
	if !exists {
		cmd.Response <- ErrSubscriptionNotExist
		return
	}

	q.AckMessages(cmd.AckIDs)

	cmd.Response <- nil
}

func (s *Scheduler) handleCreateTopic(cmd *createTopicCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if topic, exists := s.topics[cmd.TopicName]; exists {
		cmd.Response <- topic
		return
	}

	newTopic := &topic{
		subscribers: make(map[string]*queue.Queue),
	}

	s.topics[cmd.TopicName] = newTopic

	cmd.Response <- newTopic
	return
}

func (s *Scheduler) handleAddSubscriber(cmd *addSubscriberCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, exists := s.topics[cmd.TopicName]
	if !exists {
		cmd.Response <- nil
		return
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	if queue, exists := topic.subscribers[cmd.SubscriberName]; exists {
		cmd.Response <- queue
		return
	}

	newQueue := queue.NewQueue()

	topic.subscribers[cmd.SubscriberName] = newQueue

	cmd.Response <- newQueue
}

func (s *Scheduler) handlePublishMessage(cmd *publishMessageCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, exists := s.topics[cmd.TopicName]
	if !exists {
		cmd.Response <- ErrTopicNotExist
		return
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	for _, queue := range topic.subscribers {
		queue.AddMessages(cmd.Messages)
	}

	cmd.Response <- nil
	return
}

func (s *Scheduler) handleGetMessages(cmd *getMessagesCommand) {
	topic, exists := s.topics[cmd.TopicName]
	if !exists {
		cmd.Response <- nil
		return
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	q, exists := topic.subscribers[cmd.SubscriberName]
	if !exists {
		cmd.Response <- nil
		return
	}

	cmd.Response <- q.GetMessages(cmd.Count)
	return
}

func (s *Scheduler) handleGetSubscriberQueue(cmd *getSubscriberQueueCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, exists := s.topics[cmd.TopicName]
	if !exists {
		cmd.Response <- nil
		return
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	q, exists := topic.subscribers[cmd.SubscriberName]

	if !exists {
		cmd.Response <- nil
		return
	}
	cmd.Response <- q
}

func (s *Scheduler) handleRemoveSubscriber(cmd *removeSubscriberCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, exists := s.topics[cmd.TopicName]
	if !exists {
		cmd.Response <- ErrTopicNotExist
		return
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	q, exists := topic.subscribers[cmd.SubscriberName]

	q.Clear()

	delete(topic.subscribers, cmd.SubscriberName)

	cmd.Response <- nil
}

func (s *Scheduler) handleDeleteTopic(cmd *deleteTopicCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, exists := s.topics[cmd.TopicName]
	if !exists {
		cmd.Response <- ErrTopicNotExist
		return
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	for key, queue := range topic.subscribers {
		queue.Clear()
		delete(topic.subscribers, key)
	}

	delete(s.topics, cmd.TopicName)

	cmd.Response <- nil
}

func (s *Scheduler) requeueMonitor(interval time.Duration, done chan os.Signal) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				topics := s.topics
				s.mu.Unlock()

				var wg sync.WaitGroup

				for _, topic := range topics {

					topic.mu.Lock()
					subscribers := topic.subscribers
					topic.mu.Unlock()

					for _, qs := range subscribers {
						wg.Add(1)
						go func(q *queue.Queue) {
							defer wg.Done()
							q.RequeueExpiredMessages()
						}(qs)
					}
				}

				wg.Wait()
			case <-done:
				fmt.Print("stopping requeueMonitor")
				break
			}
		}
	}()
}
