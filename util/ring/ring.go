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
	"fmt"
)

// Buffer implements a circular buffer. It is a fixed size,
// and new writes overwrite older data, such that for a buffer
// of size N, for any amount of writes, only the last N bytes
// are retained.
type Buffer struct {
	data        []byte
	size        int64
	writeCursor int64
	written     int64
	newBytes    int64
}

// NewBuffer creates a new buffer of a given size.
// NewBuffer panics if the size is not greater than 0.
func NewBuffer(size int64) *Buffer {
	if size <= 0 {
		panic(fmt.Errorf("Size must be positive"))
	}

	b := &Buffer{
		size: size,
		data: make([]byte, size),
	}
	return b
}

// Write writes up to len(buf) bytes to the internal ring,
// overriding older data if necessary.
func (b *Buffer) Write(buf []byte) (int, error) {
	// Account for total bytes written
	n := len(buf)
	b.written += int64(n)
	b.newBytes += int64(n)

	// If the buffer is larger than ours, then we only care
	// about the last size bytes anyways
	if int64(n) > b.size {
		buf = buf[int64(n)-b.size:]
	}

	// Copy in place
	remain := b.size - b.writeCursor
	copy(b.data[b.writeCursor:], buf)
	if int64(len(buf)) > remain {
		copy(b.data, buf[remain:])
	}

	// Update location of the cursor
	b.writeCursor = ((b.writeCursor + int64(len(buf))) % b.size)
	return n, nil
}

// Size returns the size of the buffer
func (b *Buffer) Size() int64 {
	return b.size
}

// TotalWritten provides the total number of bytes written
func (b *Buffer) TotalWritten() int64 {
	return b.newBytes
}

// Bytes provides a slice of the bytes written. This
// slice should not be written to.
func (b *Buffer) Bytes() []byte {
	switch {
	case b.written >= b.size && b.writeCursor == 0:
		return b.data
	case b.written > b.size:
		out := make([]byte, b.size)
		copy(out, b.data[b.writeCursor:])
		copy(out[b.size-b.writeCursor:], b.data[:b.writeCursor])
		return out
	default:
		return b.data[:b.writeCursor]
	}
}

// Reset resets the buffer so it has no content.
func (b *Buffer) Reset() {
	b.writeCursor = 0
	b.written = 0
	b.newBytes = 0
}

// ResetTotalWritten resets the TotalWritten count.
func (b *Buffer) ResetTotalWritten() {
	b.newBytes = 0
}

// String returns the contents of the buffer as a string
func (b *Buffer) String() string {
	return string(b.Bytes())
}
