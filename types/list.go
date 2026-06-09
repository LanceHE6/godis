package types

import "sync"

type listNode struct {
	prev  *listNode
	next  *listNode
	value string
}

type ListValue struct {
	mu   sync.RWMutex
	head *listNode
	tail *listNode
	size int
}

func NewListValue() *ListValue {
	return &ListValue{}
}

// Load 从切片恢复链表（用于反序列化）
func (l *ListValue) Load(values []string) {
	for _, v := range values {
		node := &listNode{value: v, prev: l.tail}
		if l.tail != nil {
			l.tail.next = node
		}
		l.tail = node
		if l.head == nil {
			l.head = node
		}
		l.size++
	}
}

// Data 将链表导出为切片（用于序列化）
func (l *ListValue) Data() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]string, 0, l.size)
	node := l.head
	for node != nil {
		result = append(result, node.value)
		node = node.next
	}
	return result
}

func (l *ListValue) PushLeft(values ...string) int {
	l.mu.Lock()
	for _, v := range values {
		node := &listNode{value: v, next: l.head}
		if l.head != nil {
			l.head.prev = node
		}
		l.head = node
		if l.tail == nil {
			l.tail = node
		}
		l.size++
	}
	l.mu.Unlock()
	return l.size
}

func (l *ListValue) PushRight(values ...string) int {
	l.mu.Lock()
	for _, v := range values {
		node := &listNode{value: v, prev: l.tail}
		if l.tail != nil {
			l.tail.next = node
		}
		l.tail = node
		if l.head == nil {
			l.head = node
		}
		l.size++
	}
	l.mu.Unlock()
	return l.size
}

func (l *ListValue) PopLeft() (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.head == nil {
		return "", false
	}
	node := l.head
	l.head = node.next
	if l.head != nil {
		l.head.prev = nil
	} else {
		l.tail = nil
	}
	l.size--
	return node.value, true
}

func (l *ListValue) PopRight() (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tail == nil {
		return "", false
	}
	node := l.tail
	l.tail = node.prev
	if l.tail != nil {
		l.tail.next = nil
	} else {
		l.head = nil
	}
	l.size--
	return node.value, true
}

// Index 按下标访问，支持负索引（-1 表示最后一个）
func (l *ListValue) Index(index int) (string, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.size == 0 {
		return "", false
	}
	if index < 0 {
		index = l.size + index
	}
	if index < 0 || index >= l.size {
		return "", false
	}

	var node *listNode
	if index < l.size/2 {
		node = l.head
		for i := 0; i < index; i++ {
			node = node.next
		}
	} else {
		node = l.tail
		for i := l.size - 1; i > index; i-- {
			node = node.prev
		}
	}
	return node.value, true
}

// Range 返回 [start, stop] 范围内的元素，支持负索引
func (l *ListValue) Range(start, stop int) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.size == 0 {
		return nil
	}
	if start < 0 {
		start = l.size + start
	}
	if stop < 0 {
		stop = l.size + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= l.size {
		stop = l.size - 1
	}
	if start > stop {
		return nil
	}

	node := l.head
	for i := 0; i < start; i++ {
		node = node.next
	}

	result := make([]string, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, node.value)
		node = node.next
	}
	return result
}

func (l *ListValue) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.size
}
