package sortedmap

import (
	"sort"
)

type SortedMap struct {
	data map[string]string
	keys []string
}

func (s *SortedMap) Put(k string, v string) {
	s.data[k] = v
	i := sort.Search(len(s.keys), func(i int) bool { return s.keys[i] >= k })
	s.keys = append(s.keys, "")
	copy(s.keys[i+1:], s.keys[i:])
	s.keys[i] = k
}

func (s *SortedMap) Get(k string) string {
	return s.data[k]
}

func (s *SortedMap) Keys() []string {
	return s.keys
}

func New() *SortedMap {
	return &SortedMap{
		data: make(map[string]string),
		keys: make([]string, 0),
	}
}
