package worker

import (
	"funnel/util/ring"
	"sync"
)

func newTailer(size int64, out func(string)) (*tailer, error) {
	buf, err := ring.NewBuffer(size)
	if err != nil {
		return nil, err
	}
	return &tailer{buf: buf, out: out}, nil
}

type tailer struct {
	out func(string)
	buf *ring.Buffer
	mtx sync.Mutex
}

func (t *tailer) Write(b []byte) (int, error) {
	t.mtx.Lock()
	t.mtx.Unlock()
	w, err := t.buf.Write(b)
	if err != nil {
		return w, err
	}
	if t.buf.TotalWritten() > 100 {
		t.Flush()
	}
	return w, nil
}

func (t *tailer) Flush() {
	t.mtx.Lock()
	t.mtx.Unlock()
	if t.buf.TotalWritten() > 0 {
		t.out(t.buf.String())
		t.buf.Reset()
	}
}
