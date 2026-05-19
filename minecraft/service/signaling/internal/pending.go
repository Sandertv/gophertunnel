package internal

import (
	"sync"

	"github.com/google/uuid"
)

// NewPendingMap returns a PendingMap ready for use.
func NewPendingMap() *PendingMap {
	return &PendingMap{expected: make(map[uuid.UUID]chan error)}
}

// PendingMap tracks pending outbound messages by ID.
// Each registered ID maps to a channel that is completed when the
// corresponding message is completed.
type PendingMap struct {
	expected map[uuid.UUID]chan error
	mu       sync.Mutex
}

// Add registers interest in the completion of the outbound message with
// the given ID and returns a channel that is completed by [PendingMap.Done].
func (i *PendingMap) Add(id uuid.UUID) <-chan error {
	ch := make(chan error, 1)
	i.mu.Lock()
	i.expected[id] = ch
	i.mu.Unlock()
	return ch
}

// Remove stops tracking the outbound message with the given ID and closes
// its channel if it is still registered.
// It is typically deferred after [PendingMap.Add] once waiting for
// completion is no longer needed.
func (i *PendingMap) Remove(id uuid.UUID) {
	i.mu.Lock()
	ch, ok := i.expected[id]
	delete(i.expected, id)
	i.mu.Unlock()
	if ok {
		close(ch)
	}
}

// Done resolves the pending message registered with the given ID.
// It sends err on the associated channel, closes it, and reports whether
// the ID was registered.
func (i *PendingMap) Done(id uuid.UUID, err error) bool {
	i.mu.Lock()
	ch, ok := i.expected[id]
	delete(i.expected, id)
	i.mu.Unlock()
	if ok {
		ch <- err
		close(ch)
	}
	return ok
}
