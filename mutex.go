package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

/*
	分片map
	new MapStruct
	[]map 切片
	map 数量
	sub map
*/

// ShardMap 分片地图
type ShardMap struct {
	shards []*shard
	count  int // 分片数量
}

type shard struct {
	mu   sync.RWMutex
	data map[string]int
}

// NewShardMap 创建分片地图
func NewShardMap(shardCount int) *ShardMap {
	sm := &ShardMap{
		shards: make([]*shard, shardCount),
		count:  shardCount,
	}
	// make 初始化每个小分片
	for i := 0; i < shardCount; i++ {
		sm.shards[i] = &shard{
			data: make(map[string]int),
		}
	}
	return sm
}

// getShard 根据 key 计算所属分片
func (sm *ShardMap) getShard(key string) (*shard, int) {
	h := fnv.New32a()
	h.Write([]byte(key))
	index := int(h.Sum32() % uint32(sm.count))
	return sm.shards[index], index
}

// Set 设置值
func (sm *ShardMap) Set(key string, value int) int {
	s, index := sm.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return index
}

// Get 获取值
func (sm *ShardMap) Get(key string) (int, bool, int) {
	s, index := sm.getShard(key)
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok, index
}

// Delete 删除值
func (sm *ShardMap) Delete(key string) {
	s, _ := sm.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Count 统计总数（需要锁定所有分片）
func (sm *ShardMap) Count() int {
	total := 0
	for _, s := range sm.shards {
		s.mu.RLock()
		total += len(s.data)
		s.mu.RUnlock()
	}
	return total
}

func main() {

	sm := NewShardMap(16)
	var wg sync.WaitGroup
	// 并发写入
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", id)
			shardIndex := sm.Set(key, id)
			fmt.Printf("key_%d 写入分片: %d\n", id, shardIndex)
		}(i)
	}
	wg.Wait()
	// 验证数据
	fmt.Printf("总数: %d\n", sm.Count())
	// 读取测试
	if val, ok, shardIndex := sm.Get("key_0"); ok {
		fmt.Printf("key_0 = %d (分片: %d)\n", val, shardIndex)
	}

}
