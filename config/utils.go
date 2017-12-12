package config

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// EnsureServerProperties ensures that the server address and server password
// is consistent between the worker, node, and server.
func EnsureServerProperties(conf Config) Config {
	conf.RPC.ServerAddress = conf.Server.RPCAddress()
	conf.RPC.ServerPassword = conf.Server.Password

	conf.Node.ServerAddress = conf.Server.RPCAddress()
	conf.Node.ServerPassword = conf.Server.Password
	return conf
}

// ToYaml formats the configuration into YAML and returns the bytes.
func ToYaml(c Config) []byte {
	// TODO handle error
	yamlstr, _ := yaml.Marshal(c)
	return yamlstr
}

// ToYamlFile writes the configuration to a YAML file.
func ToYamlFile(c Config, path string) {
	// TODO handle error
	ioutil.WriteFile(path, ToYaml(c), 0600)
}

// ToYamlTempFile writes the configuration to a YAML temp. file.
func ToYamlTempFile(c Config, name string) (string, func()) {
	tmpdir, _ := ioutil.TempDir("", "")

	cleanup := func() {
		os.RemoveAll(tmpdir)
	}

	p := filepath.Join(tmpdir, name)
	ToYamlFile(c, p)
	return p, cleanup
}

// Parse parses a YAML doc into the given Config instance.
func Parse(raw []byte, conf *Config) error {
	j, err := yaml.YAMLToJSON(raw)
	if err != nil {
		return err
	}
	err = checkForUnknownKeys(j, conf)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(raw, conf)
	if err != nil {
		return err
	}
	return nil
}

// ParseFile parses a Funnel config file, which is formatted in YAML,
// and returns a Config struct.
func ParseFile(relpath string, conf *Config) error {
	if relpath == "" {
		return nil
	}

	// Try to get absolute path. If it fails, fall back to relative path.
	path, abserr := filepath.Abs(relpath)
	if abserr != nil {
		path = relpath
	}

	// Read file
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config at path %s: \n%v", path, err)
	}

	// Parse file
	err = Parse(source, conf)
	if err != nil {
		return fmt.Errorf("failed to parse config at path %s: \n%v", path, err)
	}
	return nil
}

func getKeys(obj interface{}) []string {
	keys := []string{}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			embedded := v.Type().Field(i).Anonymous
			name := v.Type().Field(i).Name
			keys = append(keys, name)

			valKeys := getKeys(field.Interface())
			vk := []string{}
			for _, v := range valKeys {
				if embedded {
					vk = append(vk, v)
				}
				vk = append(vk, name+"."+v)
			}
			keys = append(keys, vk...)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			name := key.String()
			keys = append(keys, key.String())

			valKeys := getKeys(v.MapIndex(key).Interface())
			for i, v := range valKeys {
				valKeys[i] = name + "." + v
			}
			keys = append(keys, valKeys...)
		}
	}

	return keys
}

func checkForUnknownKeys(jsonStr []byte, obj interface{}) error {
	knownMap := make(map[string]interface{})
	known := getKeys(obj)
	for _, k := range known {
		knownMap[k] = nil
	}

	var anon interface{}
	err := json.Unmarshal(jsonStr, &anon)
	if err != nil {
		return err
	}

	unknown := []string{}
	all := getKeys(anon)
	for _, k := range all {
		if _, found := knownMap[k]; !found {
			unknown = append(unknown, k)
		}
	}

	errs := []string{}
	if len(unknown) > 0 {
		for _, k := range unknown {
			parts := strings.Split(k, ".")
			field := parts[len(parts)-1]
			path := parts[:len(parts)-1]
			errs = append(
				errs,
				fmt.Sprintf("\t field %s not found in %s", field, strings.Join(path, ".")),
			)
		}
		return fmt.Errorf("%v", strings.Join(errs, "\n"))
	}

	return nil
}
