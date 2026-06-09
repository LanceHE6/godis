package types

import "sync"

type HashValue struct {
	Mu     sync.RWMutex
	Fields map[string]string
}

func NewHashValue() *HashValue {
	return &HashValue{Fields: make(map[string]string)}
}

func (h *HashValue) Set(field, value string) {
	h.Mu.Lock()
	h.Fields[field] = value
	h.Mu.Unlock()
}

func (h *HashValue) Get(field string) (string, bool) {
	h.Mu.RLock()
	v, ok := h.Fields[field]
	h.Mu.RUnlock()
	return v, ok
}

func (h *HashValue) Del(field string) bool {
	h.Mu.Lock()
	_, exists := h.Fields[field]
	if exists {
		delete(h.Fields, field)
	}
	h.Mu.Unlock()
	return exists
}

func (h *HashValue) Len() int {
	h.Mu.RLock()
	defer h.Mu.RUnlock()
	return len(h.Fields)
}

func (h *HashValue) Keys() []string {
	h.Mu.RLock()
	keys := make([]string, 0, len(h.Fields))
	for k := range h.Fields {
		keys = append(keys, k)
	}
	h.Mu.RUnlock()
	return keys
}

func (h *HashValue) Values() []string {
	h.Mu.RLock()
	vals := make([]string, 0, len(h.Fields))
	for _, v := range h.Fields {
		vals = append(vals, v)
	}
	h.Mu.RUnlock()
	return vals
}

func (h *HashValue) GetAll() map[string]string {
	h.Mu.RLock()
	result := make(map[string]string, len(h.Fields))
	for k, v := range h.Fields {
		result[k] = v
	}
	h.Mu.RUnlock()
	return result
}

func (h *HashValue) Exists(field string) bool {
	h.Mu.RLock()
	_, ok := h.Fields[field]
	h.Mu.RUnlock()
	return ok
}
