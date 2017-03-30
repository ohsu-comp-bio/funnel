package run

import (
  "bytes"
  "text/template"
  "net/url"
  "path/filepath"
  "github.com/kballard/go-shellquote"
)

// parseTpl parses a command template and extracts data into parseResult.
// "vars" contains template variable values.
func parseTpl(raw string, vars map[string]string) (*parseResult, error) {
  res := &parseResult{
    Inputs: map[string]*url.URL{},
    Outputs: map[string]*url.URL{},
  }
  funcs := template.FuncMap{
    "input": res.AddInput,
    "output": res.AddOutput,
  }

  t, err := template.New("Command").
    Delims("{", "}").
    Funcs(funcs).
    Parse(raw)

  if err != nil {
    return nil, err
  }

  buf := &bytes.Buffer{}
  xerr := t.Execute(buf, vars)
  if xerr != nil {
    return nil, xerr
  }

  cmd := buf.String()
  // TODO shell splitting needed
  res.Cmd, _ = shellquote.Split(cmd)
  return res, nil
}

// parseResult captures data extracted from parsing the command template.
type parseResult struct {
  Inputs map[string]*url.URL
  Outputs map[string]*url.URL
  Cmd []string
}

func (res *parseResult) AddInput(rawurl string) string {
  u := parseUrl(rawurl)
  p := "/opt/funnel/inputs" + u.EscapedPath()
  res.Inputs[p] = u
  return p
}

func (res *parseResult) AddOutput(rawurl string) string {
  u := parseUrl(rawurl)
  p := "/opt/funnel/outputs" + u.EscapedPath()
  res.Outputs[p] = u
  return p
}

func parseUrl(rawurl string) *url.URL {
  u, _ := url.Parse(rawurl)
  if u.Scheme == "" {
    u.Scheme = "file"
    abs, _ := filepath.Abs(filepath.Clean(rawurl))
    u, _ = url.Parse(abs)
  }
  return u
}
