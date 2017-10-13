package events

// Writer provides write access to a task's events
type Writer interface {
	Write(*Event) error
	Close() error
}

type multiwriter []Writer

// MultiWriter writes events to all the given writers.
func MultiWriter(ws ...Writer) Writer {
	return multiwriter(ws)
}

// Write writes an event to all the writers.
func (mw multiwriter) Write(ev *Event) error {
	for _, w := range mw {
		err := w.Write(ev)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mw multiwriter) Close() error {
	for _, w := range mw {
		w.Close()
	}
	return nil
}
