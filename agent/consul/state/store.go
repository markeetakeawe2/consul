package state

import (
	sync
)

type Store struct {
	mu       sync.RWMutex
	services map[string]bool
}

func NewStore() *Store {
	return &Store{
	services: make(map[string]bool),
	}
}

func (s *Store) Register(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[id] = true
}

func (s *Store) Deregister(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, id)
}

func (s *Store) IsRegistered(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.services[id]
}

func (s *Store) Snapshot() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap := make(map[string]bool)
	for k, v := range s.services {
		snap[k] = v
	}
	return snap
}

func (s *Store) Restore(snap map[string]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services = make(map[string]bool)
	for k, v := range snap {
		s.services[k] = v
	}
}
