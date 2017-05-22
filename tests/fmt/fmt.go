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

	// Call `go test` and pass all the arguments.
	goPath, _ := exec.LookPath("go")
	cmd := exec.Command(goPath, append([]string{"test"}, os.Args[1:]...)...)
	cmdStdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	scanner := bufio.NewScanner(cmdStdout)

	// Collect a report that can be dumped at the end.
	report := new(bytes.Buffer)
	out := io.MultiWriter(report, os.Stdout)
	// Don't print the output summary default, because in many cases
	// it will just look like the go test output was printed twice
	// (e.g. if there were no logs to clog it all up).
	printSummary := false

	cmd.Start()

	// Scan the `go test` output looking for report lines which we can colorize.
	for scanner.Scan() {
		b := scanner.Bytes()
		s := string(b)

		switch {
		// PASS lines
		case passtail.Match(b) || passlonely.Match(b):
			fmt.Fprintf(out, "%s\n", aurora.Green(s))

		// RUN, SKIP, ok, and lots of other stuff from `go test`
		case runhead.Match(b) || skiptail.Match(b) || okPath.Match(b) ||
			notestPath.Match(b) || coverage.Match(b) || filename.Match(b) ||
			importpath.Match(b):

			fmt.Fprintf(out, "%s\n", aurora.Colorize(s, 257))

		// FAIL
		case failtail.Match(b):
			fmt.Fprintf(out, "%s\n", aurora.Red(s))

		// Everything else, most likely funnel logging.
		default:
			// Only print the summary if there were lines that didn't match
			// known go test output (most likely logs)
			printSummary = true
			fmt.Fprintln(os.Stdout, s)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "scanning stdout: ", err)
	}

	cmderr := cmd.Wait()

	if printSummary {
		fmt.Println("\n=========== Summary ============")
		fmt.Printf(report.String())
	}

	if cmderr != nil {
		os.Exit(1)
	}
}
