package types

import "sync"

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

// RandomMembers 随机返回 count 个成员（count > 0 可重复，< 0 不可重复）
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

	if count > 0 {
		result := make([]string, count)
		for i := 0; i < count; i++ {
			result[i] = members[i%len(members)]
		}
		return result
	}

	if -count > len(members) {
		count = -len(members)
	}
	result := make([]string, -count)
	copy(result, members[:-count])
	return result
}
