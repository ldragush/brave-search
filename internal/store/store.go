package store

import "sync"

type ResultStore struct {
	mu   sync.Mutex
	set  map[string]struct{}
}

func NewResultStore() *ResultStore {
	return &ResultStore{
		set: make(map[string]struct{}, 4096),
	}
}

// Add returns true if the value was new.
func (s *ResultStore) Add(v string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.set[v]; ok {
		return false
	}
	s.set[v] = struct{}{}
	return true
}

func (s *ResultStore) Values() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.set))
	for k := range s.set {
		out = append(out, k)
	}
	return out
}
