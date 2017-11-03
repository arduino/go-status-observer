package statesync

import (
	"sync"
)

// Observer is an objects that can receive status updates
type Observer interface {
	OnUpdate(newstate interface{}) error
	ReceiveChan() <-chan interface{}
}

// State is a status container tha can send updates on Observers
type State struct {
	status    interface{}
	observers []Observer
	recvQ     chan interface{}
	lock      sync.Mutex
}

// UpdateStatus sends status update to all observers
func (s *State) UpdateStatus(status interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.status = status

	// If an observer is not valid, remove it from the array
	i := 0
	for _, observer := range s.observers {
		if err := observer.OnUpdate(status); err == nil {
			s.observers[i] = observer
			i++
		}
	}
	s.observers = s.observers[:i]
}

// AddObserver adds an observer to the list of status observers.
// A one-shot status update is sent immediately as the observer
// is added to put the observer in sync with the others.
func (s *State) AddObserver(observer Observer) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := observer.OnUpdate(s.status); err != nil {
		// failed to add observer
		return
	}

	if s.recvQ == nil {
		s.recvQ = make(chan interface{})
	}
	go receiveFeeder(s.recvQ, observer.ReceiveChan())
	s.observers = append(s.observers, observer)
}

func receiveFeeder(out chan<- interface{}, in <-chan interface{}) {
	for {
		data, ok := <-in
		if !ok {
			return
		}
		out <- data
	}
}

// Receive returns messages coming back from observers.
// The messages are received without any ordering, it may be needed
// to call the method several times to retrieve all the queued messages.
func (s *State) Receive() interface{} {
	s.lock.Lock()
	if s.recvQ == nil {
		s.recvQ = make(chan interface{})
	}
	s.lock.Unlock()

	res, ok := <-s.recvQ
	if ok {
		return res
	}
	return nil
}
