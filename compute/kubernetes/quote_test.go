package kubernetes

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// minimalExecutorTemplate mirrors the args section of executor-job.yaml so
// that quote-handling can be tested without a live cluster.
const minimalExecutorTemplate = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{.TaskId}}-{{.JobId}}
  namespace: default
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: exec
        image: alpine
        command: ["/bin/sh", "-c"]
        args:
          {{- range .Command}}
          - {{.}}
          {{- end}}
`

// yamlQuote mirrors the production helper in worker/kubernetes.go.
func yamlQuote(s string) string {
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return "\"" + escaped + "\""
}

func renderArgs(t *testing.T, command []string) []string {
	t.Helper()

	quoted := make([]string, len(command))
	for i, v := range command {
		quoted[i] = yamlQuote(v)
	}

	tpl, err := template.New("exec").Parse(minimalExecutorTemplate)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	data := map[string]interface{}{
		"TaskId":  "task-1",
		"JobId":   0,
		"Command": quoted,
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatalf("execute template: %v", err)
	}

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(buf.Bytes(), nil, nil)
	if err != nil {
		t.Fatalf("decode rendered YAML: %v\n---\n%s", err, buf.String())
	}

	job, ok := obj.(*batchv1.Job)
	if !ok {
		t.Fatalf("decoded object is not a Job: %T", obj)
	}

	containers := job.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		t.Fatal("no containers in rendered job")
	}
	return containers[0].Args
}

func TestQuoteHandling(t *testing.T) {
	tests := []struct {
		name     string
		command  []string
		wantArgs []string
	}{
		{
			name:     "plain words",
			command:  []string{"echo", "hello"},
			wantArgs: []string{"echo", "hello"},
		},
		{
			name:     "single quote in argument",
			command:  []string{"echo", "Hello O'hare!"},
			wantArgs: []string{"echo", "Hello O'hare!"},
		},
		{
			name:     "double quote in argument",
			command:  []string{"echo", `"double quoted"`},
			wantArgs: []string{"echo", `"double quoted"`},
		},
		{
			name:     "mixed quotes",
			command:  []string{"echo", `"mix 'of' quotes"`},
			wantArgs: []string{"echo", `"mix 'of' quotes"`},
		},
		{
			name:     "curl with header — full command as single element",
			command:  []string{`curl -s https://api.github.com/zen -H "Accept: application/json,text/event-stream"`},
			wantArgs: []string{`curl -s https://api.github.com/zen -H "Accept: application/json,text/event-stream"`},
		},
		{
			name:     "curl with header — properly split arguments",
			command:  []string{"curl", "-s", "https://api.github.com/zen", "-H", "Accept: application/json,text/event-stream"},
			wantArgs: []string{"curl", "-s", "https://api.github.com/zen", "-H", "Accept: application/json,text/event-stream"},
		},
		{
			name:     "shell inline with single quotes",
			command:  []string{"sh", "-c", "echo 'inline shell' && ls -1"},
			wantArgs: []string{"sh", "-c", "echo 'inline shell' && ls -1"},
		},
		{
			name:     "argument with backslash",
			command:  []string{"echo", `back\slash`},
			wantArgs: []string{"echo", `back\slash`},
		},
		{
			name:     "empty argument preserved",
			command:  []string{"echo", ""},
			wantArgs: []string{"echo", ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := renderArgs(t, tc.command)
			if len(got) != len(tc.wantArgs) {
				t.Fatalf("args length mismatch: got %d (%v), want %d (%v)", len(got), got, len(tc.wantArgs), tc.wantArgs)
			}
			for i := range tc.wantArgs {
				if got[i] != tc.wantArgs[i] {
					t.Errorf("args[%d]: got %q, want %q", i, got[i], tc.wantArgs[i])
				}
			}
		})
	}
}
