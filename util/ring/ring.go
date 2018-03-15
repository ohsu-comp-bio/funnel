package ring

import (
	"github.com/armon/circbuf"
)

// Buffer implements a circular buffer. It is a fixed size,
// and new writes overwrite older data, such that for a buffer
// of size N, for any amount of writes, only the last N bytes
// are retained.
type Buffer struct {
	*circbuf.Buffer
	newBytes int64
}

// NewBuffer creates a new buffer of a given size.
// NewBuffer panics if the size is not greater than 0.
func NewBuffer(size int64) *Buffer {
	buff, err := circbuf.NewBuffer(size)
	if err != nil {
		panic(err)
	}

	b := &Buffer{
		Buffer: buff,
	}
	return b
}

// Write writes up to len(buf) bytes to the internal ring,
// overriding older data if necessary.
func (b *Buffer) Write(buf []byte) (int, error) {
	n := len(buf)
	b.newBytes += int64(n)
	return b.Buffer.Write(buf)
}

// NewBytesWritten provides the total number of bytes written since the last reset.
func (b *Buffer) NewBytesWritten() int64 {
	return b.newBytes
}

// Reset resets the buffer so it has no content.
func (b *Buffer) Reset() {
	b.Buffer.Reset()
	b.newBytes = 0
}

// ResetNewBytesWritten resets the new bytes written counter.
func (b *Buffer) ResetNewBytesWritten() {
	b.newBytes = 0
}
