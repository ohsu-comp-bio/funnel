package worker

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

type multiWriter []TaskWriter

func (m multiWriter) State(s tes.State) {
	for _, w := range m {
		w.State(s)
	}
}

func (m multiWriter) StartTime(t time.Time) {
	for _, w := range m {
		w.StartTime(t)
	}
}

func (m multiWriter) EndTime(t time.Time) {
	for _, w := range m {
		w.EndTime(t)
	}
}

func (m multiWriter) Outputs(o []*tes.OutputFileLog) {
	for _, w := range m {
		w.Outputs(o)
	}
}

func (m multiWriter) Metadata(meta map[string]string) {
	for _, w := range m {
		w.Metadata(meta)
	}
}

func (m multiWriter) ExecutorExitCode(i int, code int) {
	for _, w := range m {
		w.ExecutorExitCode(i, code)
	}
}

func (m multiWriter) ExecutorPorts(i int, ports []*tes.Ports) {
	for _, w := range m {
		w.ExecutorPorts(i, ports)
	}
}

func (m multiWriter) ExecutorHostIP(i int, ip string) {
	for _, w := range m {
		w.ExecutorHostIP(i, ip)
	}
}

func (m multiWriter) ExecutorStartTime(i int, t time.Time) {
	for _, w := range m {
		w.ExecutorStartTime(i, t)
	}
}

func (m multiWriter) ExecutorEndTime(i int, t time.Time) {
	for _, w := range m {
		w.ExecutorEndTime(i, t)
	}
}

func (m multiWriter) AppendExecutorStdout(i int, s string) {
	for _, w := range m {
		w.AppendExecutorStdout(i, s)
	}
}

func (m multiWriter) AppendExecutorStderr(i int, s string) {
	for _, w := range m {
		w.AppendExecutorStderr(i, s)
	}
}
