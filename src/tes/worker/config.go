package tesTaskEngineWorker

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	pbr "tes/server/proto"
)

// Config contains worker configuration.
type Config struct {
	ID string
	// Address of the scheduler, e.g. "1.2.3.4:9090"
	ServerAddress string
	// Directory to write job files to
	WorkDir    string
	Timeout    int
	NumWorkers int
	Storage    []*pbr.StorageConfig
	LogPath    string
}

// DefaultConfig returns simple, default worker configuration.
func DefaultConfig() Config {
	return Config{
		ServerAddress: "localhost:9090",
		WorkDir:       "tes-work-dir",
		Timeout:       -1,
		NumWorkers:    4,
		LogPath:       "tes-worker-log",
	}
}

// ToYaml formats the configuration into YAML and returns the bytes.
func (c Config) ToYaml() []byte {
	// TODO handle error
	yamlstr, _ := yaml.Marshal(c)
	return yamlstr
}

// ToYamlFile writes the configuration to a YAML file.
func (c Config) ToYamlFile(p string) {
	// TODO handle error
	ioutil.WriteFile(p, c.ToYaml(), 0600)
}

// ToYamlTempFile writes the configuration to a YAML temp. file.
func (c Config) ToYamlTempFile(name string) (string, func()) {
	// I'm creating a temp. directory instead of a temp. file so that
	// the file can have an expected name. This is helpful for the HTCondor scheduler.
	tmpdir, _ := ioutil.TempDir("", "")

	cleanup := func() {
		os.RemoveAll(tmpdir)
	}

	p := filepath.Join(tmpdir, name)
	c.ToYamlFile(p)
	return p, cleanup
}
