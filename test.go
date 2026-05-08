package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

type BoaMapShard struct {
	count     int
	boaShards []*BoaShard
}

type BoaShard struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewBoaMapShard(count int) *BoaMapShard {
	sm := &BoaMapShard{
		count:     count,
		boaShards: make([]*BoaShard, count),
	}
	for i := 0; i < count; i++ {
		sm.boaShards[i] = &BoaShard{
			data: make(map[string]interface{}),
		}
	}
	return sm
}

// 获取所属分片
func (sm *BoaMapShard) getOwnShard(key string) (*BoaShard, int) {
	h := fnv.New32a()
	h.Write([]byte(key))
	index := int(h.Sum32() % uint32(sm.count))
	return sm.boaShards[index], index
}

func (sm *BoaMapShard) Set(key string, value interface{}) {
	s, _ := sm.getOwnShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (sm *BoaMapShard) Get(key string) (interface{}, bool) {
	s, _ := sm.getOwnShard(key)
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (sm *BoaMapShard) Count() int {
	total := 0
	for _, s := range sm.boaShards {
		s.mu.RLock()
		total += len(s.data)
		s.mu.RUnlock()
	}
	return total
}

func main() {

	sm := NewBoaMapShard(6)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", id)
			//fmt.Println("写入 === ", id)
			sm.Set(key, id)
		}(i)
	}
	wg.Wait()
	fmt.Println("总数 === ", sm.Count())

	if val, ok := sm.Get("key_5"); ok {
		fmt.Println("读取 === ", val)
	}

}
