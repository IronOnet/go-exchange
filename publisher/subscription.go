package publisher

import (
	"sync"
)

type Subscription struct {
	Subscribers map[string]map[int64]*Client
	Mu          sync.RWMutex
}

func NewSubscription() *Subscription {
	return &Subscription{Subscribers: map[string]map[int64]*Client{}}
}

func (s *Subscription) Subscribe(channel string, client *Client) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	_, found := s.Subscribers[channel]
	if !found {
		s.Subscribers[channel] = map[int64]*Client{}
	}

	_, found = s.Subscribers[channel][client.Id]
	if found {
		return false
	}
	s.Subscribers[channel][client.Id] = client
	return true
}

func (s *Subscription) Unsubscribe(channel string, client *Client) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	_, found := s.Subscribers[channel]
	if !found {
		return false
	}

	_, found = s.Subscribers[channel][client.Id]
	if !found {
		return false
	}

	delete(s.Subscribers[channel], client.Id)
	return true
}

func (s *Subscription) Publish(channel string, msg interface{}) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	_, found := s.Subscribers[channel]
	if !found {
		return
	}

	for _, c := range s.Subscribers[channel] {
		c.WriteCh <- msg
	}
}
