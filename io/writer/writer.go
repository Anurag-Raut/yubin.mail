package writer

import (
	"bufio"
	"fmt"
	"net"
)

type Writer struct {
	*bufio.Writer
}

func NewWriter(conn net.Conn) *Writer {
	return &Writer{Writer: bufio.NewWriter(conn)}
}

func (w *Writer) WriteString(s string) (n int, err error) {
	n, err = w.Writer.WriteString(s)
	if err != nil {
		return n, err
	}
	err = w.Flush()

	return n, err
}

func (w *Writer) Fprintf(format string, args ...any) error {
	_, err := fmt.Fprintf(w.Writer, format, args...)
	if err != nil {
		return err
	}
	return w.Flush()
}
