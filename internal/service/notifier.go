package service

import (
	"encoding/json"
	"sync"
)

type NotificationEvent string

const (
	EventFlightContainmentInfringement NotificationEvent = "flight_containment.infringed"
	EventTacticalConflictDetected     NotificationEvent = "tactical_conflict.detected"
)

type eventMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Notifier struct {
	mu          sync.RWMutex
	subscribers map[NotificationEvent]map[*subscriber]struct{}
}

type subscriber struct {
	ch chan []byte
}

func NewNotifier() *Notifier {
	return &Notifier{
		subscribers: make(map[NotificationEvent]map[*subscriber]struct{}),
	}
}

func (n *Notifier) Subscribe(event NotificationEvent) (<-chan []byte, func()) {
	sub := &subscriber{ch: make(chan []byte, 16)}
	n.mu.Lock()
	if n.subscribers[event] == nil {
		n.subscribers[event] = make(map[*subscriber]struct{})
	}
	n.subscribers[event][sub] = struct{}{}
	n.mu.Unlock()
	return sub.ch, func() {
		n.mu.Lock()
		subs := n.subscribers[event]
		if subs != nil {
			if _, ok := subs[sub]; ok {
				delete(subs, sub)
				if len(subs) == 0 {
					delete(n.subscribers, event)
				}
			}
		}
		n.mu.Unlock()
		close(sub.ch)
	}
}

func (n *Notifier) Publish(event NotificationEvent, payload interface{}) error {
	msg := eventMessage{
		Type: string(event),
		Data: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	n.mu.RLock()
	subs := n.subscribers[event]
	list := make([]*subscriber, 0, len(subs))
	for sub := range subs {
		list = append(list, sub)
	}
	n.mu.RUnlock()
	for _, sub := range list {
		select {
		case sub.ch <- data:
		default:
		}
	}
	return nil
}
