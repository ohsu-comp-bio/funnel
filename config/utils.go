package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

// ToYaml formats the configuration into YAML and returns the bytes.
func ToYaml(c *Config) ([]byte, error) {
	return yaml.Marshal(c)
}

// ToYamlFile writes the configuration to a YAML file.
func ToYamlFile(c *Config, path string) error {
	b, err := ToYaml(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

// Parse parses a YAML doc into the given Config instance.
func Parse(raw []byte, conf *Config) error {
	if err := yaml.UnmarshalWithOptions(raw, conf); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Field-level value validation
	validate := validator.New()
	if err := validate.Struct(conf); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
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
	source, err := os.ReadFile(path)
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
