package tes


import (
  "tes/server/proto"
  "github.com/ghodss/yaml"
  "io/ioutil"
)

func ParseConfigFile(path string) (ga4gh_task_ref.ServerConfig, error) {
  doc := ga4gh_task_ref.ServerConfig{}
  source, err := ioutil.ReadFile(path)
  	if err != nil {
  		return doc, err
	}
  err = yaml.Unmarshal(source, &doc)
  return doc, nil
}