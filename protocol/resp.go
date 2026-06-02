package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseRESP 从reader中解析命令
func ParseRESP(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil, fmt.Errorf("收到空行")
	}

	if line[0] != '*' {
		return nil, fmt.Errorf("非法协议: 期待以 '*' 开头，实际收到 '%c'", line[0])
	}

	arrayLen, err := strconv.Atoi(line[1:])
	if err != nil || arrayLen <= 0 {
		return nil, fmt.Errorf("非法的数组长度")
	}

	args := make([]string, 0, arrayLen)

	for i := 0; i < arrayLen; i++ {
		strLenLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		strLenLine = strings.TrimSpace(strLenLine)

		if len(strLenLine) == 0 || strLenLine[0] != '$' {
			return nil, fmt.Errorf("非法协议: 期待以 '$' 开头")
		}

		strLen, err := strconv.Atoi(strLenLine[1:])
		if err != nil {
			return nil, fmt.Errorf("非法的 Bulk String 长度")
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
