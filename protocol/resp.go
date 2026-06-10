package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseRESP 从 reader 中解析命令，支持 RESP 协议和 inline 模式
func ParseRESP(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil, fmt.Errorf("received an empty line")
	}

	// RESP 协议：以 * 开头的数组格式
	if line[0] == '*' {
		return parseRESPArray(line, reader)
	}

	// inline 模式：按空格拆分（redis-cli 等客户端在部分场景下使用）
	return parseInline(line), nil
}

// parseRESPArray 解析 RESP 多条批量数组
func parseRESPArray(firstLine string, reader *bufio.Reader) ([]string, error) {
	arrayLen, err := strconv.Atoi(firstLine[1:])
	if err != nil || arrayLen <= 0 {
		return nil, fmt.Errorf("invalid array length")
	}

	args := make([]string, 0, arrayLen)

	for i := 0; i < arrayLen; i++ {
		strLenLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		strLenLine = strings.TrimSpace(strLenLine)

		if len(strLenLine) == 0 || strLenLine[0] != '$' {
			return nil, fmt.Errorf("invalid protocol")
		}

		strLen, err := strconv.Atoi(strLenLine[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length")
		}

		buf := make([]byte, strLen)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			return nil, err
		}

		args = append(args, string(buf))
		_, _ = reader.ReadString('\n')
	}

	return args, nil
}

// parseInline 解析 inline 命令（如 "SET key value"）
func parseInline(line string) []string {
	return strings.Fields(line)
}
