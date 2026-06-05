package protocol

import (
	"strconv"
)

// MakeSimpleString 构造简单字符串 (如: +OK\r\n)
func MakeSimpleString(s string) string {
	return "+" + s + "\r\n"
}

// MakeError 构造错误回复 (如: -ERR message\r\n)
func MakeError(err string) string {
	return "-" + err + "\r\n"
}

// MakeInt 构造整数回复 (如: :100\r\n)
func MakeInt(n int) string {
	return ":" + strconv.Itoa(n) + "\r\n"
}

// MakeBulkString 构造大块字符串 (如: $5\r\nhello\r\n)
func MakeBulkString(s string) string {
	return "$" + strconv.Itoa(len(s)) + "\r\n" + s + "\r\n"
}

// MakeNull 构造空回复 (如: $-1\r\n)
func MakeNull() string {
	return "$-1\r\n"
}

// MakeArray 构造数组，接收已经序列化好的元素的切片
func MakeArray(respElements []string) string {
	res := "*" + strconv.Itoa(len(respElements)) + "\r\n"
	for _, el := range respElements {
		res += el
	}
	return res
}
