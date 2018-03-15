package ring

// Copied and modified from: https://github.com/armon/circbuf
/*
The MIT License (MIT)

Copyright (c) 2013 Armon Dadgar

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

import (
	"bytes"
	"io"
	"testing"
)

func TestBuffer_Impl(t *testing.T) {
	var _ io.Writer = &Buffer{}
}

func TestBuffer_ShortWrite(t *testing.T) {
	buf := NewBuffer(1024)
	inp := []byte("hello world")

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("bad: %v", n)
	}

	if !bytes.Equal(buf.Bytes(), inp) {
		t.Fatalf("bad: %v", buf.Bytes())
	}
}

func TestBuffer_FullWrite(t *testing.T) {
	inp := []byte("hello world")
	buf := NewBuffer(int64(len(inp)))

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("bad: %v", n)
	}

	if !bytes.Equal(buf.Bytes(), inp) {
		t.Fatalf("bad: %v", buf.Bytes())
	}
}

func TestBuffer_LongWrite(t *testing.T) {
	inp := []byte("hello world")
	buf := NewBuffer(6)

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("bad: %v", n)
	}

	expect := []byte(" world")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: %s", buf.Bytes())
	}
}

func TestBuffer_HugeWrite(t *testing.T) {
	inp := []byte("hello world")
	buf := NewBuffer(3)

	n, err := buf.Write(inp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != len(inp) {
		t.Fatalf("bad: %v", n)
	}

	expect := []byte("rld")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: %s", buf.Bytes())
	}
}

func TestBuffer_ManySmall(t *testing.T) {
	inp := []byte("hello world")
	buf := NewBuffer(3)

	for _, b := range inp {
		n, err := buf.Write([]byte{b})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if n != 1 {
			t.Fatalf("bad: %v", n)
		}
	}

	expect := []byte("rld")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: %v", buf.Bytes())
	}
}

func TestBuffer_MultiPart(t *testing.T) {
	inputs := [][]byte{
		[]byte("hello world\n"),
		[]byte("this is a test\n"),
		[]byte("my cool input\n"),
	}
	total := 0

	buf := NewBuffer(16)

	for _, b := range inputs {
		total += len(b)
		n, err := buf.Write(b)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if n != len(b) {
			t.Fatalf("bad: %v", n)
		}
	}

	if int64(total) != buf.TotalWritten() {
		t.Fatalf("bad total")
	}

	expect := []byte("t\nmy cool input\n")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: %v", buf.Bytes())
	}
}

func TestBuffer_Reset(t *testing.T) {
	// Write a bunch of data
	inputs := [][]byte{
		[]byte("hello world\n"),
		[]byte("this is a test\n"),
		[]byte("my cool input\n"),
	}

	buf := NewBuffer(4)

	for _, b := range inputs {
		n, err := buf.Write(b)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if n != len(b) {
			t.Fatalf("bad: %v", n)
		}
	}

	// Reset it
	buf.Reset()

	// Write more data
	input := []byte("hello")
	n, err := buf.Write(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if n != len(input) {
		t.Fatalf("bad: %v", n)
	}

	// Test the output
	expect := []byte("ello")
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Fatalf("bad: %v", string(buf.Bytes()))
	}
}

func TestBuffer_ResetTotalWritten(t *testing.T) {

	buf := NewBuffer(4)
	buf.Write([]byte("12345"))

	if buf.TotalWritten() != 5 {
		t.Fatalf("expected total written to be 4, got %d", buf.TotalWritten())
	}

	if buf.String() != "2345" {
		t.Fatalf("expected content to be 2345, got %s", buf.String())
	}
	buf.ResetTotalWritten()

	if buf.TotalWritten() != 0 {
		t.Fatalf("expected total written to be 0, got %d", buf.TotalWritten())
	}

	if buf.String() != "2345" {
		t.Fatalf("expected content to be 2345, got %s", buf.String())
	}
}
