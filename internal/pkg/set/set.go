package set

import (
	"fmt"
	"sync"
)

type Set struct {
	set   map[string]struct{}
	mutex *sync.Mutex
}

func NewSet(items []string) (a Set) {
	a = Set{
		set:   make(map[string]struct{}),
		mutex: &sync.Mutex{},
	}

	for _, item := range items {
		_ = a.Add(item)
	}

	return
}

func (s Set) Add(v string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.set[v]; ok {
		return fmt.Errorf("value %s already in Set", v)
	}

	// Empty struct does not consume mem
	s.set[v] = struct{}{}

	return nil
}

func (s *Set) Values() (keys []string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	keys = make([]string, 0, len(s.set))
	for k := range s.set {
		keys = append(keys, k)
	}

	return
}

func (s *Set) Has(v string) (ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok = s.set[v]

	return
}

func (s *Set) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.set = make(map[string]struct{})
}