package main

import "sync"

type jobStorage struct {
	m sync.Mutex
	v map[string]int
}

func newJobStorage() jobStorage {
	return jobStorage{v: make(map[string]int)}
}

func (s *jobStorage) Set(key string, pid int) {
	s.m.Lock()
	defer s.m.Unlock()

	s.v[key] = pid
}

func (s *jobStorage) Get(key string) (pid int, prs bool) {
	s.m.Lock()
	defer s.m.Unlock()

	pid, prs = s.v[key]
	return
}

func (s *jobStorage) Remove(key string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.v, key)
}
