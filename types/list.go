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

// Set 设置指定位置的值
func (l *ListValue) Set(index int, value string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if index < 0 {
		index = l.size + index
	}
	if index < 0 || index >= l.size {
		return false
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
	node.value = value
	return true
}

// InsertBefore 在 pivot 之前插入 value，返回新长度；未找到 pivot 返回 -1
func (l *ListValue) InsertBefore(pivot, value string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	for node := l.head; node != nil; node = node.next {
		if node.value == pivot {
			n := &listNode{value: value, prev: node.prev, next: node}
			if node.prev != nil {
				node.prev.next = n
			} else {
				l.head = n
			}
			node.prev = n
			l.size++
			return l.size
		}
	}
	return -1
}

// InsertAfter 在 pivot 之后插入 value，返回新长度；未找到 pivot 返回 -1
func (l *ListValue) InsertAfter(pivot, value string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	for node := l.head; node != nil; node = node.next {
		if node.value == pivot {
			n := &listNode{value: value, prev: node, next: node.next}
			if node.next != nil {
				node.next.prev = n
			} else {
				l.tail = n
			}
			node.next = n
			l.size++
			return l.size
		}
	}
	return -1
}

// Remove 删除 count 个值为 value 的节点，count>0 从头删，count<0 从尾删，count=0 删全部
// 返回实际删除的数量
func (l *ListValue) Remove(value string, count int) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.size == 0 {
		return 0
	}

	removed := 0
	if count == 0 {
		// 删除全部匹配的节点
		node := l.head
		for node != nil {
			next := node.next
			if node.value == value {
				l.removeNode(node)
				removed++
			}
			node = next
		}
	} else if count > 0 {
		// 从头删 count 个
		node := l.head
		for node != nil && removed < count {
			next := node.next
			if node.value == value {
				l.removeNode(node)
				removed++
			}
			node = next
		}
	} else {
		// 从尾删 |count| 个
		count = -count
		node := l.tail
		for node != nil && removed < count {
			prev := node.prev
			if node.value == value {
				l.removeNode(node)
				removed++
			}
			node = prev
		}
	}
	return removed
}

// removeNode 从链表中移除节点（内部使用，调用方需持有锁）
func (l *ListValue) removeNode(node *listNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}
	l.size--
}

// Find 查找元素位置，从 start 开始，跳过 skip 个匹配，返回 index（-1 未找到）
func (l *ListValue) Find(value string, start, skip int) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if start < 0 {
		start = l.size + start
	}
	if start < 0 {
		start = 0
	}
	node := l.head
	for i := 0; i < start && node != nil; i++ {
		node = node.next
	}
	for idx := start; node != nil; idx++ {
		if node.value == value {
			if skip == 0 {
				return idx
			}
			skip--
		}
		node = node.next
	}
	return -1
}
