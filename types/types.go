package types

// DataType 数据类型枚举
type DataType int

const (
	TypeString DataType = iota // 0
	TypeHash                   // 1
	TypeList                   // 2
	TypeSet                    // 3
	TypeZSet                   // 4
)
