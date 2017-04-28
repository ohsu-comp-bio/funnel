package run

import (
	"errors"
	set "github.com/deckarep/golang-set"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"regexp"
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
	keys := []set.Set{}
	for i, mymap := range maps {
		keys = append(keys, set.NewSet())
		for k := range mymap {
			if keys[i].Contains(k) {
				err := errors.New("Can't use the same KEY for multiple --in, --out, --env arguments: " + k)
				return err
			}
			keys[i].Add(k)
		}
	}

	common := set.NewSet()
	i := 0
	for i < (len(keys) - 1) {
		j := i + 1
		for j <= (len(keys) - 1) {
			common = common.Union(keys[i].Intersect(keys[j]))
			j++
		}
		i++
	}

	if common.Cardinality() > 0 {
		err := errors.New("Can't use the same KEY for multiple --in, --out, --env arguments: " + common.String())
		return err
	}
	return nil
}

func fileMapToEnvVars(m map[string]string, path string) (map[string]string, error) {
	result := map[string]string{}
	for k, v := range m {
		p, err := stripStoragePrefix(v)
		if err != nil {
			return nil, err
		}
		result[k] = path + p
	}
	return result, nil
}

func createTaskParams(params map[string]string, path string, t tes.FileType) ([]*tes.TaskParameter, error) {
	result := []*tes.TaskParameter{}
	for key, url := range params {
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
