package scheduler

import (
	"errors"

	"github.com/axelburling/dfs/pkg/pubsub/apiv1/message"
	"github.com/axelburling/dfs/pkg/pubsub/apiv1/scheduler/queue"
)

var (
	ErrTopicCreate        = errors.New("Could not create Topic")
	ErrSubscriptionCreate = errors.New("Could not create Subscription")
	ErrSubscriptionGet    = errors.New("Could not get Subscription")
	ErrMessagesGet        = errors.New("Could not get Messages")
)

func (s *Scheduler) CreateTopic(topicName string) (*topic, error) {
	responseChan := make(chan *topic)
	s.commandChannel <- &createTopicCommand{
		TopicName: topicName,
		Response:  responseChan,
	}

	topic := <-responseChan
	close(responseChan)

	if topic == nil {
		return nil, ErrTopicCreate
	}

	return topic, nil
}

func (s *Scheduler) DeleteTopic(topicName string) error {
	responseChan := make(chan error)
	s.commandChannel <- &deleteTopicCommand{
		TopicName: topicName,
		Response:  responseChan,
	}

	err := <-responseChan
	close(responseChan)

	return err
}

func (s *Scheduler) CreateSubscription(topicName, subName string) (*queue.Queue, error) {
	responseChan := make(chan *queue.Queue)

	s.commandChannel <- &addSubscriberCommand{
		TopicName:      topicName,
		SubscriberName: subName,
		Response:       responseChan,
	}

	queue := <-responseChan
	close(responseChan)

	if queue == nil {
		return nil, ErrSubscriptionCreate
	}

	return queue, nil
}

func (s *Scheduler) RemoveSubscription(topicName, subName string) error {
	responseChan := make(chan error)
	s.commandChannel <- &removeSubscriberCommand{
		TopicName:      topicName,
		SubscriberName: subName,
		Response:       responseChan,
	}

	err := <-responseChan
	close(responseChan)

	return err
}

func (s *Scheduler) Publish(topicName string, msgs []*message.Message) error {

	responseChan := make(chan error)
	s.commandChannel <- &publishMessageCommand{
		TopicName: topicName,
		Messages:  msgs,
		Response:  responseChan,
	}

	err := <-responseChan
	close(responseChan)

	return err
}

func (s *Scheduler) GetMessages(topicName, subName string, count int) ([]*message.Message, error) {
	responseChan := make(chan []*message.Message)
	s.commandChannel <- &getMessagesCommand{
		TopicName:      topicName,
		SubscriberName: subName,
		Count:          count,
		Response:       responseChan,
	}

	messages := <-responseChan
	// close(responseChan)

	if messages == nil {
		return nil, ErrMessagesGet
	}

	return messages, nil
}

func (s *Scheduler) GetSubscription(topicName, subName string) (*queue.Queue, error) {
	responseChan := make(chan *queue.Queue)

	s.commandChannel <- &getSubscriberQueueCommand{
		TopicName:      topicName,
		SubscriberName: subName,
		Response:       responseChan,
	}

	queue := <-responseChan
	close(responseChan)

	if queue == nil {
		return nil, ErrSubscriptionGet
	}

	return queue, nil
}

func (s *Scheduler) GetTopic(topicName string) (*topic, error) {
	responseChan := make(chan *topic)

	s.commandChannel <- &getTopicCommand{
		TopicName: topicName,
		Response:  responseChan,
	}

	t := <-responseChan
	close(responseChan)

	if t == nil {
		return nil, ErrTopicNotExist
	}

	return t, nil
}

func (s *Scheduler) AckMessages(topicName, subName string, ids []string) error {
	responseChan := make(chan error)
	s.commandChannel <- &ackMessagesCommand{
		TopicName:      topicName,
		SubscriberName: subName,
		AckIDs:         ids,
		Response:       responseChan,
	}

	err := <-responseChan
	close(responseChan)

	return err
}
