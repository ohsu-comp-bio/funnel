package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io"
	"os"
	"regexp"
	"strings"
)

var Done = fmt.Errorf("iterator done")

var address = "10.96.11.130"

func main() {
	ctx := context.Background()

	// TODO byte progress would be nice
	r := NewSyslogEventReader(os.Stdin)

	for {
		ev, err := r.Next()
		if err == Done {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(ev.Msg)

		task := tes.Task{
			Id: ev.TaskID,
		}

		up := client.Update().
			Index("funnel").
			Type("task").
			Id(ev.TaskID).
			Doc(task).
			Refresh("true").
			DocAsUpsert(true)

		if _, err := up.Do(ctx); err != nil {
			fmt.Println("Err", task)
		} else {
			fmt.Println("Write", task)
		}
	}
}

type SyslogEventReader struct {
	scan *bufio.Scanner
}

func NewSyslogEventReader(r io.Reader) *SyslogEventReader {
	return &SyslogEventReader{bufio.NewScanner(r)}
}

//var rx = regexp.MustCompile(`^\w+\s+\d+\s+\d+:\d+:\d+\s+(.+)\s+(.+)(\[\d+\])?:\s+(.*)`)
var rx = regexp.MustCompile(`^[a-zA-Z]+\s+(\d+)\s+(\d+):(\d+):(\d+)\s+(\S+)\s+(\S+):\s+(.*)`)

func (s *SyslogEventReader) Next() (*logmsg, error) {
	for s.scan.Scan() {

		m := rx.FindSubmatch(s.scan.Bytes())
		if m != nil {
			app := string(m[6])
			blob := string(m[7])

			if strings.HasPrefix(app, "funnel") {
				if ev := readTaskLog(blob); ev != nil {
					return ev, nil
				}
			}
		}
	}
	if err := s.scan.Err(); err != nil {
		return nil, err
	}
	return nil, Done
}

func readTaskLog(raw string) *logmsg {

	m := logmsg{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		fmt.Println("skipping:", raw)
		return nil
	}
	return &m
}

const (
	StateChange         = "Set task state"
	TaskStartTime       = "StartTime"
	TaskEndTime         = "EndTime"
	TaskOutputs         = "Outputs"
	TaskMetadata        = "Metadata"
	ExecutorExitCode    = "ExecutorExitCode"
	ExecutorHostIP      = "ExecutorHostIP"
	ExecutorStartTime   = "ExecutorStartTime"
	ExecutorEndTime     = "ExecutorEndTime"
	ExecutorPorts       = "ExecutorPorts"
	ExecutorStdoutChunk = "AppendExecutorStdout"
	ExecutorStderrChunk = "AppendExecutorStderr"
)

type logmsg struct {
	Msg    string
	TaskID string
	State  string

	StartTime string
	EndTime   string

	ExecutorIndex        int
	ExecutorStartTime    string
	ExecutorEndTime      string
	ExecutorHostIP       string
	AppendExecutorStdout string
	AppendExecutorStderr string
	ExecutorExitCode     int32

	ExecutorPorts []*tes.Ports
	Outputs       []*tes.OutputFileLog
	Metadata      map[string]string
}
