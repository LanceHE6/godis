package types

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

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

func (z *ZSetValue) Load(members []ZSetMember) {
	for _, m := range members {
		z.Scores[m.Member] = m.Score
	}
	z.Dirty = true
}

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

// RangeRev 返回降序排名（-1 为最高分）
func (z *ZSetValue) RangeRev(start, stop int) []ZSetMember {
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
	for i := stop; i >= start; i-- {
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

// RevRank 返回倒序排名（从高到低，0 是最高分），不存在返回 -1
func (z *ZSetValue) RevRank(member string) int {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()

	size := len(z.Sorted)
	for i := size - 1; i >= 0; i-- {
		if z.Sorted[i] == member {
			return size - 1 - i
		}
	}
	return -1
}

// RangeByScore 返回 score 在 [min, max] 范围内的成员
func (z *ZSetValue) RangeByScore(min, max float64, offset, count int) []ZSetMember {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()

	skipped := 0
	collected := 0
	var result []ZSetMember
	for _, m := range z.Sorted {
		s := z.Scores[m]
		if s < min {
			continue
		}
		if s > max {
			break
		}
		skipped++
		if offset > 0 && skipped <= offset {
			continue
		}
		if count > 0 && collected >= count {
			break
		}
		result = append(result, ZSetMember{Member: m, Score: s})
		collected++
	}
	return result
}

// CountByScore 统计 [min, max] 范围内的成员数
func (z *ZSetValue) CountByScore(min, max float64) int {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()

	count := 0
	for _, m := range z.Sorted {
		s := z.Scores[m]
		if s < min {
			continue
		}
		if s > max {
			break
		}
		count++
	}
	return count
}

// CountByLex 统计字典序在 [min, max] 范围内的成员数
func (z *ZSetValue) CountByLex(min, max string, minEx, maxEx bool) int {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()

	count := 0
	for _, m := range z.Sorted {
		if !minEx && m < min {
			continue
		}
		if minEx && m <= min {
			continue
		}
		if !maxEx && m > max {
			break
		}
		if maxEx && m >= max {
			break
		}
		count++
	}
	return count
}

// RangeByLex 字典序范围查询，支持 LIMIT
func (z *ZSetValue) RangeByLex(min, max string, minEx, maxEx bool, offset, count int) []string {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()

	skipped := 0
	collected := 0
	result := make([]string, 0)
	for _, m := range z.Sorted {
		if !minEx && m < min {
			continue
		}
		if minEx && m <= min {
			continue
		}
		if !maxEx && m > max {
			break
		}
		if maxEx && m >= max {
			break
		}
		skipped++
		if offset > 0 && skipped <= offset {
			continue
		}
		if count > 0 && collected >= count {
			break
		}
		result = append(result, m)
		collected++
	}
	return result
}

// IncrBy 增加 member 的分数，返回新分数；不存在时创建
func (z *ZSetValue) IncrBy(increment float64, member string) float64 {
	z.Mu.Lock()
	z.Scores[member] += increment
	z.Dirty = true
	newScore := z.Scores[member]
	z.Mu.Unlock()
	return newScore
}

// PopMin 弹出 count 个最低分成员
func (z *ZSetValue) PopMin(count int) []ZSetMember {
	z.Mu.Lock()
	defer z.Mu.Unlock()
	z.EnsureSorted()

	if count > len(z.Sorted) {
		count = len(z.Sorted)
	}
	result := make([]ZSetMember, 0, count)
	for i := 0; i < count; i++ {
		m := z.Sorted[i]
		result = append(result, ZSetMember{Member: m, Score: z.Scores[m]})
		delete(z.Scores, m)
	}
	z.Sorted = z.Sorted[count:]
	if len(z.Sorted) == 0 {
		z.Sorted = nil
	}
	return result
}

// PopMax 弹出 count 个最高分成员
func (z *ZSetValue) PopMax(count int) []ZSetMember {
	z.Mu.Lock()
	defer z.Mu.Unlock()
	z.EnsureSorted()

	size := len(z.Sorted)
	if count > size {
		count = size
	}
	result := make([]ZSetMember, 0, count)
	for i := size - 1; i >= size-count; i-- {
		m := z.Sorted[i]
		result = append(result, ZSetMember{Member: m, Score: z.Scores[m]})
		delete(z.Scores, m)
	}
	z.Sorted = z.Sorted[:size-count]
	return result
}

// RemoveByScore 移除 [min, max] 范围内的成员，返回数量
func (z *ZSetValue) RemoveByScore(min, max float64) int {
	z.Mu.Lock()
	defer z.Mu.Unlock()
	z.EnsureSorted()

	removed := 0
	newSorted := make([]string, 0, len(z.Sorted))
	for _, m := range z.Sorted {
		s := z.Scores[m]
		if s >= min && s <= max {
			delete(z.Scores, m)
			removed++
		} else {
			newSorted = append(newSorted, m)
			if s > max {
				break // 已超过范围，后面都不需要
			}
		}
	}
	if removed > 0 {
		// 追加剩余元素
		for i := len(newSorted); i < len(z.Sorted); i++ {
			if z.Scores[z.Sorted[i]] > max {
				newSorted = append(newSorted, z.Sorted[i])
			}
		}
		z.Sorted = newSorted
	}
	return removed
}

// RemoveByRank 移除 [start, stop] 排名范围内的成员
func (z *ZSetValue) RemoveByRank(start, stop int) int {
	z.Mu.Lock()
	defer z.Mu.Unlock()
	z.EnsureSorted()

	size := len(z.Sorted)
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
	if start > stop || start >= size {
		return 0
	}

	removed := 0
	for i := start; i <= stop; i++ {
		delete(z.Scores, z.Sorted[i])
		removed++
	}
	z.Sorted = append(z.Sorted[:start], z.Sorted[stop+1:]...)
	z.Dirty = false
	return removed
}

// RemoveByLex 移除字典序 [min, max] 范围内的成员
func (z *ZSetValue) RemoveByLex(min, max string, minEx, maxEx bool) int {
	z.Mu.Lock()
	defer z.Mu.Unlock()
	z.EnsureSorted()

	removed := 0
	newSorted := make([]string, 0, len(z.Sorted))
	for _, m := range z.Sorted {
		belowMin := minEx && m <= min || !minEx && m < min
		aboveMax := maxEx && m >= max || !maxEx && m > max

		if belowMin {
			newSorted = append(newSorted, m)
		} else if aboveMax {
			newSorted = append(newSorted, m)
		} else {
			delete(z.Scores, m)
			removed++
		}
	}
	z.Sorted = newSorted
	return removed
}

// RandomMembers 随机返回 count 个成员（正=不重复，负=可重复）
func (z *ZSetValue) RandomMembers(count int) []ZSetMember {
	z.Mu.RLock()
	defer z.Mu.RUnlock()

	if len(z.Scores) == 0 {
		return nil
	}

	members := make([]string, 0, len(z.Scores))
	for m := range z.Scores {
		members = append(members, m)
	}

	if count < 0 {
		result := make([]ZSetMember, -count)
		for i := range result {
			m := members[rand.Intn(len(members))]
			result[i] = ZSetMember{Member: m, Score: z.Scores[m]}
		}
		return result
	}

	if count > len(members) {
		count = len(members)
	}
	// Fisher-Yates
	for i := 0; i < count; i++ {
		j := i + rand.Intn(len(members)-i)
		members[i], members[j] = members[j], members[i]
	}
	result := make([]ZSetMember, count)
	for i := 0; i < count; i++ {
		result[i] = ZSetMember{Member: members[i], Score: z.Scores[members[i]]}
	}
	return result
}

// Pop 随机弹出 count 个成员
func (z *ZSetValue) Pop(count int) []ZSetMember {
	z.Mu.Lock()
	defer z.Mu.Unlock()

	if len(z.Scores) == 0 {
		return nil
	}

	members := make([]string, 0, len(z.Scores))
	for m := range z.Scores {
		members = append(members, m)
	}

	if count > len(members) {
		count = len(members)
	}

	// Fisher-Yates partial shuffle
	result := make([]ZSetMember, count)
	for i := 0; i < count; i++ {
		j := i + rand.Intn(len(members)-i)
		members[i], members[j] = members[j], members[i]
		result[i] = ZSetMember{Member: members[i], Score: z.Scores[members[i]]}
		delete(z.Scores, members[i])
	}
	z.Dirty = true
	return result
}

// Members 返回所有成员（用于 ZSCAN/ZDIFF/ZUNION/ZINTER）
func (z *ZSetValue) Members() []string {
	z.Mu.RLock()
	defer z.Mu.RUnlock()
	z.EnsureSorted()
	m := make([]string, len(z.Sorted))
	copy(m, z.Sorted)
	return m
}

// ---- score 边界解析 ----

// ScoreBound 解析后的分数边界
type ScoreBound struct {
	Value     float64
	Exclusive bool
	Infinite  bool // 表示 -inf 或 +inf
}

// ParseScoreBound 解析 "-inf", "+inf", "(1.5", "5" 等格式
func ParseScoreBound(s string) (ScoreBound, error) {
	switch strings.ToLower(s) {
	case "+inf":
		return ScoreBound{Value: math.Inf(1), Infinite: true}, nil
	case "-inf":
		return ScoreBound{Value: math.Inf(-1), Infinite: true}, nil
	}
	excl := false
	ns := s
	if strings.HasPrefix(s, "(") {
		excl = true
		ns = s[1:]
	}
	v, err := strconv.ParseFloat(ns, 64)
	if err != nil {
		return ScoreBound{}, err
	}
	return ScoreBound{Value: v, Exclusive: excl}, nil
}
