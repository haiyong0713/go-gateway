package dynamicV2

import (
	"sync"

	"github.com/pkg/errors"
)

type Set struct {
	sync.RWMutex
	set map[interface{}]struct{}
}

func (s *Set) Len() int {
	return len(s.set)
}

func (s *Set) Add(items ...interface{}) {
	for _, item := range items {
		s.set[item] = struct{}{}
	}
}

func (s *Set) AddSync(items ...interface{}) {
	s.Lock()
	defer s.Unlock()
	s.Add(items...)
}

func (s *Set) Exists(item interface{}) bool {
	_, ok := s.set[item]
	return ok
}

func (s *Set) Delete(items ...interface{}) {
	for _, item := range items {
		delete(s.set, item)
	}
}

func (s *Set) DeleteSync(items ...interface{}) {
	s.Lock()
	defer s.Unlock()
	s.Delete(items...)
}

func (s *Set) GetList() []interface{} {
	var ret []interface{}
	s.RLock()
	defer s.RUnlock()
	for k := range s.set {
		ret = append(ret, k)
	}
	return ret
}

func (s *Set) GetListToInt64() ([]int64, error) {
	var ret []int64
	s.RLock()
	defer s.RUnlock()
	for k := range s.set {
		switch item := k.(type) {
		case int64:
			ret = append(ret, item)
		default:
			return nil, errors.Errorf("Cannot convert %T to int64\n", item)
		}
	}
	return ret, nil
}

func (s *Set) GetListToString() ([]string, error) {
	var ret []string
	s.RLock()
	defer s.RUnlock()
	for k := range s.set {
		switch item := k.(type) {
		case string:
			ret = append(ret, item)
		default:
			return nil, errors.Errorf("Cannot convert %T to string\n", item)
		}
	}
	return ret, nil
}
