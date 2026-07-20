package types

import (
	"math/rand"
	"sync"
)

type SetValue struct {
	Mu      sync.RWMutex
	Members map[string]struct{}
}

func NewSetValue() *SetValue {
	return &SetValue{Members: make(map[string]struct{})}
}

func (s *SetValue) Add(members ...string) int {
	s.Mu.Lock()
	added := 0
	for _, m := range members {
		if _, exists := s.Members[m]; !exists {
			s.Members[m] = struct{}{}
			added++
		}
	}
	s.Mu.Unlock()
	return added
}

func (s *SetValue) Remove(members ...string) int {
	s.Mu.Lock()
	removed := 0
	for _, m := range members {
		if _, exists := s.Members[m]; exists {
			delete(s.Members, m)
			removed++
		}
	}
	s.Mu.Unlock()
	return removed
}

func (s *SetValue) Exists(member string) bool {
	s.Mu.RLock()
	_, ok := s.Members[member]
	s.Mu.RUnlock()
	return ok
}

func (s *SetValue) MembersList() []string {
	s.Mu.RLock()
	result := make([]string, 0, len(s.Members))
	for m := range s.Members {
		result = append(result, m)
	}
	s.Mu.RUnlock()
	return result
}

func (s *SetValue) Len() int {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	return len(s.Members)
}

// RandomMembers 随机返回 count 个成员（count >= 0 不可重复，< 0 可重复）
func (s *SetValue) RandomMembers(count int) []string {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	if len(s.Members) == 0 {
		return nil
	}

	members := make([]string, 0, len(s.Members))
	for m := range s.Members {
		members = append(members, m)
	}

	if count < 0 {
		// 负数表示可重复
		result := make([]string, -count)
		for i := range result {
			result[i] = members[rand.Intn(len(members))]
		}
		return result
	}

	if count > len(members) {
		count = len(members)
	}

	// Fisher-Yates 部分洗牌
	for i := 0; i < count; i++ {
		j := i + rand.Intn(len(members)-i)
		members[i], members[j] = members[j], members[i]
	}
	result := make([]string, count)
	copy(result, members[:count])
	return result
}

// Pop 随机弹出 count 个成员并返回（count=1 时返回 []string{member} 或 nil）
func (s *SetValue) Pop(count int) []string {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if len(s.Members) == 0 {
		return nil
	}

	members := make([]string, 0, len(s.Members))
	for m := range s.Members {
		members = append(members, m)
	}

	if count > len(members) {
		count = len(members)
	}

	result := make([]string, count)
	for i := 0; i < count; i++ {
		j := i + rand.Intn(len(members)-i)
		members[i], members[j] = members[j], members[i]
		result[i] = members[i]
		delete(s.Members, members[i])
	}
	return result
}

// Intersect 计算交集，返回共同成员
func (s *SetValue) Intersect(others ...*SetValue) []string {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	if len(s.Members) == 0 {
		return nil
	}

	result := make([]string, 0)
	for m := range s.Members {
		inAll := true
		for _, o := range others {
			if !o.Exists(m) {
				inAll = false
				break
			}
		}
		if inAll {
			result = append(result, m)
		}
	}
	return result
}

// IntersectLen 返回交集大小（不列出成员）
func (s *SetValue) IntersectLen(others ...*SetValue) int {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	if len(s.Members) == 0 {
		return 0
	}

	count := 0
	for m := range s.Members {
		inAll := true
		for _, o := range others {
			if !o.Exists(m) {
				inAll = false
				break
			}
		}
		if inAll {
			count++
		}
	}
	return count
}

// Union 计算并集，返回所有唯一成员
func (s *SetValue) Union(others ...*SetValue) []string {
	seen := make(map[string]struct{})

	s.Mu.RLock()
	for m := range s.Members {
		seen[m] = struct{}{}
	}
	s.Mu.RUnlock()

	for _, o := range others {
		if o == nil {
			continue
		}
		o.Mu.RLock()
		for m := range o.Members {
			seen[m] = struct{}{}
		}
		o.Mu.RUnlock()
	}

	result := make([]string, 0, len(seen))
	for m := range seen {
		result = append(result, m)
	}
	return result
}

// Diff 计算差集，返回存在于 s 但不存在于 others 的成员
func (s *SetValue) Diff(others ...*SetValue) []string {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	result := make([]string, 0)
	for m := range s.Members {
		ok := true
		for _, o := range others {
			if o == nil {
				continue
			}
			if o.Exists(m) {
				ok = false
				break
			}
		}
		if ok {
			result = append(result, m)
		}
	}
	return result
}
