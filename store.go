package main

import (
	"errors"
	"sort"
	"sync"
)

var ErrKeyExists = errors.New("key already exists")

type Flag struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

type Store struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

func NewStore() *Store {
	return &Store{flags: make(map[string]Flag)}
}

func (s *Store) Put(f Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[f.Key]; ok {
		return ErrKeyExists
	}
	s.flags[f.Key] = f
	return nil
}

func (s *Store) Get(key string) (Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[key]
	return f, ok
}

func (s *Store) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

func (s *Store) Update(key string, f Flag) (Flag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return Flag{}, false
	}
	f.Key = key
	s.flags[key] = f
	return f, true
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return false
	}
	delete(s.flags, key)
	return true
}
