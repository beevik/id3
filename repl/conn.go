package main

import (
	"bufio"
	"fmt"
	"io"
)

type conn struct {
	input  *bufio.Scanner
	output *bufio.Writer
}

func newConn(r io.Reader, w io.Writer) *conn {
	return &conn{
		input:  bufio.NewScanner(r),
		output: bufio.NewWriter(w),
	}
}

func (c *conn) Flush() {
	c.output.Flush()
}

func (c *conn) Print(args ...interface{}) {
	fmt.Fprint(c.output, args...)
}

func (c *conn) Printf(format string, args ...interface{}) {
	fmt.Fprintf(c.output, format, args...)
	c.Flush()
}

func (c *conn) Println(args ...interface{}) {
	fmt.Fprintln(c.output, args...)
	c.Flush()
}

func (c *conn) GetLine() (string, error) {
	if c.input.Scan() {
		return c.input.Text(), nil
	}

	return "", c.input.Err()
}
