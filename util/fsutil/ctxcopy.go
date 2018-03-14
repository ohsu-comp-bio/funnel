package fsutil

import (
	"context"
	"io"
)

// Copy copies from src to dst until either the context is canceled,
// EOF is reached on src or an error occurs.
func Copy(ctx context.Context, dst io.Writer, src io.Reader) (written int64, err error) {
	return copyBufferWithContext(ctx, dst, src, nil)
}

func copyBufferWithContext(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	if buf == nil {
		buf = make([]byte, size)
	}
L:
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break L
		default:
			nr, er := src.Read(buf)
			if nr > 0 {
				nw, ew := dst.Write(buf[0:nr])
				if nw > 0 {
					written += int64(nw)
				}
				if ew != nil {
					err = ew
					break L
				}
				if nr != nw {
					err = io.ErrShortWrite
					break L
				}
			}
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break L
			}
		}
	}
	return written, err
}
