package trace

import (
	"bufio"
	"bytes"
	"io"
)

type Reader struct {
	input   io.Reader
	scanner *bufio.Scanner
}

func NewReader(input io.Reader) *Reader {
	scanner := bufio.NewScanner(input)

	buffer := make([]byte, 10<<20)
	scanner.Buffer(buffer, 1<<20)
	scanner.Split(splitStack)

	return &Reader{
		input:   input,
		scanner: scanner,
	}
}

func (reader *Reader) Read() (Event, error) {
tryagain:
	if !reader.scanner.Scan() {
		return Event{}, io.EOF
	}

	blocktext := reader.scanner.Text()
	event, ok := ParseEvent(blocktext)
	if !ok {
		goto tryagain
	}

	return event, nil
}

func splitStack(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\n', '\n'}); i >= 0 {
		return i + 2, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
