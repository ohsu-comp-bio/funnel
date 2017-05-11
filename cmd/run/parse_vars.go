package run

import (
	"errors"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"path/filepath"
	"regexp"
	"strings"
)

func mergeVars(maps ...map[string]string) (map[string]string, error) {
	var err error
	merged := map[string]string{}
	for _, m := range maps {
		err = mergo.MergeWithOverwrite(&merged, m)
		if err != nil {
			return nil, err
		}
	}
	return merged, nil
}

func parseCliVars(args []string) (map[string]string, error) {
	data := map[string]string{}

	if len(args) == 0 {
		return data, nil
	}

	for _, arg := range args {
		re := regexp.MustCompile("=")
		res := re.Split(arg, -1)
		if len(res) != 2 {
			err := errors.New("Arguments passed to --in, --out and --env must be of the form: KEY=VALUE")
			return data, err
		}
		key := res[0]
		val := res[1]
		if _, ok := data[key]; ok {
			err := errors.New("Can't use the same KEY for multiple --in, --out, --env arguments: " + key)
			return data, err
		}
		data[key] = val
	}
	return data, nil
}

func compareKeys(maps ...map[string]string) error {
	keys := make(map[string]string)
	for _, mymap := range maps {
		for k := range mymap {
			if _, ok := keys[k]; !ok {
				keys[k] = ""
			} else {
				err := errors.New("Can't use the same KEY for multiple --in, --out, --env arguments: " + k)
				return err
			}
		}
	}
	return nil
}

func stripStoragePrefix(url string) (string, error) {
	re := regexp.MustCompile("[a-z0-9]+://")
	if !re.MatchString(url) {
		err := errors.New("File paths must be prefixed with one of:\n file://\n gs://\n s3://")
		return "", err
	}
	path := re.ReplaceAllString(url, "")
	return strings.TrimPrefix(path, "/"), nil
}

func resolvePath(url string) (string, error) {
	local := strings.HasPrefix(url, "/") || strings.HasPrefix(url, ".") || strings.HasPrefix(url, "~")
	re := regexp.MustCompile("[a-z0-9]+://")
	prefixed := re.MatchString(url)
	var path string
	switch {
	case local:
		path, err := filepath.Abs(url)
		if err != nil {
			return "", err
		}
		return "file://" + path, nil
	case prefixed:
		path = url
		return path, nil
	default:
		e := fmt.Sprintf("could not resolve filepath: %s", url)
		return "", errors.New(e)
	}
}

func fileMapToEnvVars(m map[string]string, path string) (map[string]string, error) {
	result := map[string]string{}
	for k, v := range m {
		url, err := resolvePath(v)
		if err != nil {
			return nil, err
		}
		p, err := stripStoragePrefix(url)
		if err != nil {
			return nil, err
		}
		result[k] = path + p
	}
	return result, nil
}

func createTaskParams(params map[string]string, path string, t tes.FileType) ([]*tes.TaskParameter, error) {
	result := []*tes.TaskParameter{}
	for key, val := range params {
		url, err := resolvePath(val)
		if err != nil {
			return nil, err
		}
		p, err := stripStoragePrefix(url)
		if err != nil {
			return nil, err
		}
		path := path + p
		param := &tes.TaskParameter{
			Name: key,
			Url:  url,
			Path: path,
			Type: t,
		}
		result = append(result, param)
	}
	return result, nil
}
