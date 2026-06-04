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
		return nil, fmt.Errorf("received an empty line")
	}

	if line[0] != '*' {
		return nil, fmt.Errorf("invalid protocol")
	}

	arrayLen, err := strconv.Atoi(line[1:])
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
