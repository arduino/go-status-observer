package statesync

import (
	"sync"
)

// Observer is an objects that can receive status updates
type Observer interface {
	OnUpdate(newstate interface{}) error
}

// State is a status container tha can send updates on Observers
type State struct {
	status    interface{}
	observers []Observer
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
	s.observers = append(s.observers, observer)
}
