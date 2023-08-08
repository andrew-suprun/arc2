package stream

import (
	"sync"
)

type Stream[T any] struct {
	name     string
	elements []T
	closed   bool
	*sync.Cond
}

func NewStream[T any](name string) *Stream[T] {
	return &Stream[T]{
		Cond: sync.NewCond(&sync.Mutex{}),
		name: name,
	}
}

func (s *Stream[T]) Push(msg T) {
	s.Cond.L.Lock()
	defer s.Cond.L.Unlock()
	if s.closed {
		panic("Stream.Push() on closed stream")
	}
	s.elements = append(s.elements, msg)
	s.Cond.Signal()
}

func (s *Stream[T]) Pull() ([]T, bool) {
	for {
		s.Cond.L.Lock()
		if len(s.elements) == 0 && !s.closed {
			s.Cond.Wait()
			s.Cond.L.Unlock()
			continue
		} else {
			msgs := s.elements
			s.elements = []T{}
			closed := s.closed
			s.Cond.L.Unlock()
			return msgs, closed
		}
	}
}

func (s *Stream[T]) TryPull() ([]T, bool) {
	s.Cond.L.Lock()
	defer s.Cond.L.Unlock()
	msgs := s.elements
	s.elements = []T{}
	return msgs, s.closed
}

func (s *Stream[T]) Close() {
	s.Cond.L.Lock()
	s.closed = true
	s.Cond.Signal()
	s.Cond.L.Unlock()
}
