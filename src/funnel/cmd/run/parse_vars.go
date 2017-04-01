package run

import (
  "fmt"
  "github.com/imdario/mergo"
  "strings"
)

func mergeVars(a, b map[string]string) (map[string]string, error) {
  var err error
  data := map[string]string{}
  err = mergo.MergeWithOverwrite(&data, a)
  if err != nil {
    return nil, err
  }
  err = mergo.MergeWithOverwrite(&data, b)
  if err != nil {
    return nil, err
  }
  return data, nil
}

func parseFileVars(path string) (map[string]string, error) {
  data := map[string]string{}
  return data, nil
}

func parseCliVars(args []string) (map[string]string, error) {
  data := map[string]string{}

  if len(args) == 0 {
    return data, nil
  }

  key := ""
  mode := "key"

  for _, arg := range args {
    if mode == "key" {
      if !strings.HasPrefix(arg, "-") {
        return nil, fmt.Errorf("Unexpected value. Expected key (e.g. '-key' or '--key')")
      }
      key = strings.TrimLeft(arg, "-")
      mode = "value"
    } else {
      if strings.HasPrefix(arg, "-") {
        return nil, fmt.Errorf("Unexpected key. Expected value for key '%s'", key)
      }
      data[key] = arg
      mode = "key"
    }
  }

  if mode == "value" {
    return nil, fmt.Errorf("No value found for key: '%s'", key)
  }

  return data, nil
}
