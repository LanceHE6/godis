package types

type StringValue struct {
	Value string
}

func NewStringValue(val string) *StringValue {
	return &StringValue{Value: val}
}
