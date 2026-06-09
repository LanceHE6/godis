package types

import "sync"

// ZSetMember 用于批量操作的成员结构
type ZSetMember struct {
	Member string
	Score  float64
}

type ZSetValue struct {
	Mu     sync.RWMutex
	Scores map[string]float64
	Sorted []string // 按 score 升序排列，惰性排序
	Dirty  bool
}

func NewZSetValue() *ZSetValue {
	return &ZSetValue{Scores: make(map[string]float64)}
}

// Load 从切片恢复（用于反序列化）
func (z *ZSetValue) Load(members []ZSetMember) {
	for _, m := range members {
		z.Scores[m.Member] = m.Score
	}
	z.Dirty = true
}

// Data 导出为有序切片（用于序列化）
func (z *ZSetValue) Data() []ZSetMember {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()
	result := make([]ZSetMember, 0, len(z.Sorted))
	for _, m := range z.Sorted {
		result = append(result, ZSetMember{Member: m, Score: z.Scores[m]})
	}
	return result
}

func (z *ZSetValue) Add(score float64, member string) int {
	z.Mu.Lock()
	_, exists := z.Scores[member]
	z.Scores[member] = score
	z.Dirty = true
	z.Mu.Unlock()
	if exists {
		return 0
	}
	return 1
}

func (z *ZSetValue) AddMultiple(members ...ZSetMember) int {
	z.Mu.Lock()
	added := 0
	for _, m := range members {
		_, exists := z.Scores[m.Member]
		z.Scores[m.Member] = m.Score
		if !exists {
			added++
		}
	}
	z.Dirty = true
	z.Mu.Unlock()
	return added
}

func (z *ZSetValue) Remove(member string) bool {
	z.Mu.Lock()
	_, exists := z.Scores[member]
	if exists {
		delete(z.Scores, member)
		z.Dirty = true
	}
	z.Mu.Unlock()
	return exists
}

func (z *ZSetValue) Score(member string) (float64, bool) {
	z.Mu.RLock()
	s, ok := z.Scores[member]
	z.Mu.RUnlock()
	return s, ok
}

func (z *ZSetValue) Len() int {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	return len(z.Scores)
}

func (z *ZSetValue) Exists(member string) bool {
	z.Mu.RLock()
	_, ok := z.Scores[member]
	z.Mu.RUnlock()
	return ok
}

// EnsureSorted 在读取前确保排序（调用者需持锁）
func (z *ZSetValue) EnsureSorted() {
	if !z.Dirty {
		return
	}
	z.Sorted = make([]string, 0, len(z.Scores))
	for m := range z.Scores {
		z.Sorted = append(z.Sorted, m)
	}
	for i := 1; i < len(z.Sorted); i++ {
		key := z.Sorted[i]
		keyScore := z.Scores[key]
		j := i - 1
		for j >= 0 {
			jMember := z.Sorted[j]
			jScore := z.Scores[jMember]
			if jScore > keyScore || (jScore == keyScore && jMember > key) {
				z.Sorted[j+1] = z.Sorted[j]
				j--
			} else {
				break
			}
		}
		z.Sorted[j+1] = key
	}
	z.Dirty = false
}

// Range 返回排名在 [start, stop] 的成员（从 0 开始，支持负索引）
func (z *ZSetValue) Range(start, stop int) []ZSetMember {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()

	size := len(z.Sorted)
	if size == 0 {
		return nil
	}
	if start < 0 {
		start = size + start
	}
	if stop < 0 {
		stop = size + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= size {
		stop = size - 1
	}
	if start > stop {
		return nil
	}

	result := make([]ZSetMember, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		m := z.Sorted[i]
		result = append(result, ZSetMember{Member: m, Score: z.Scores[m]})
	}
	return result
}

// Rank 返回成员的排名（从 0 开始），不存在返回 -1
func (z *ZSetValue) Rank(member string) int {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()

	for i, m := range z.Sorted {
		if m == member {
			return i
		}
	}
	return -1
}

// RangeByScore 返回 score 在 [min, max] 范围内的成员
func (z *ZSetValue) RangeByScore(min, max float64) []ZSetMember {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()

	var result []ZSetMember
	for _, m := range z.Sorted {
		s := z.Scores[m]
		if s >= min && s <= max {
			result = append(result, ZSetMember{Member: m, Score: s})
		} else if s > max {
			break
		}
	}
	return result
}
