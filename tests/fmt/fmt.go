// This provides a simple program which transforms the output of `go test`
// in order to make it more readable.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/logrusorgru/aurora"
	"io"
	"os"
	"os/exec"
	"regexp"
)

var (
	runhead    = regexp.MustCompile(`(?m)^=== RUN   Test.*`)
	passtail   = regexp.MustCompile(`(?m)^(\s*)--- PASS: Test`)
	skiptail   = regexp.MustCompile(`(?m)^(\s*)--- SKIP: Test`)
	failtail   = regexp.MustCompile(`(?m)^(\s*)--- FAIL: Test`)
	passlonely = regexp.MustCompile(`(?m)^PASS$`)
	faillonely = regexp.MustCompile(`(?m)^FAIL$`)

	okPath     = regexp.MustCompile(`(?m)^ok\s+(\S+)\s+([\d\.]+\w+)(?:  (coverage: \d+\.\d+% of statements))?$`)
	failPath   = regexp.MustCompile(`(?m)^FAIL\s+\S+\s+(?:[\d\.]+\w+|\[build failed\])$`)
	notestPath = regexp.MustCompile(`(?m)^\?\s+\S+\s+\[no test files\]$`)

	coverage = regexp.MustCompile(`(?m)^coverage: ((\d+)\.\d)+% of statements?$`)

	filename   = regexp.MustCompile(`(?m)([^\s:]+\.go):(\d+)`)
	emptyline  = regexp.MustCompile(`(?m)^\s*\r?\n`)
	importpath = regexp.MustCompile(`(?m)^# (.*)$`)
)

func main() {
	goPath, err := exec.LookPath("go")
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(goPath, append([]string{"test"}, os.Args[1:]...)...)

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	cmd.Start()

	// Collect a report that can be dumped at the end.
	report := new(bytes.Buffer)
	out := io.MultiWriter(report, os.Stdout)

	go func() {
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			fmt.Fprintf(os.Stderr, "    %s\n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "scanning stderr: ", err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			b := scanner.Bytes()
			s := string(b)

			switch {
			case passtail.Match(b):
				fmt.Fprintf(out, "%s\n", aurora.Green(s))
			case runhead.Match(b):
				fmt.Fprintf(out, "%s\n", aurora.Colorize(s, 257))
			case failtail.Match(b):
				fmt.Fprintf(out, "%s\n", aurora.Red(s))
			default:
				fmt.Fprintln(out, aurora.Colorize(scanner.Text(), 257))
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "scanning stdout: ", err)
		}
	}()

	cmd.Wait()
	fmt.Println("\n=========== Summary ============")
	fmt.Printf(report.String())
}
