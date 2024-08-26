package mpsd

import "github.com/google/uuid"

type Handler interface {
	HandleSessionChange(ref SessionReference, branch uuid.UUID, changeNumber uint64)
}

type NopHandler struct{}

func (NopHandler) HandleSessionChange(SessionReference, uuid.UUID, uint64) {}

func (s *Session) Handle(h Handler) {
	if h == nil {
		h = NopHandler{}
	}
	s.h.Store(&h)
}

func (s *Session) handler() Handler {
	return *s.h.Load()
}
