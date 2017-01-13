package tesTaskEngineWorker

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	pbr "tes/server/proto"
)

// Worker configuration.
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

func DefaultConfig() Config {
	return Config{
		ServerAddress: "localhost:9090",
		WorkDir:       "tes-work-dir",
		Timeout:       -1,
		NumWorkers:    4,
		LogPath:       "tes-worker-log",
	}
}

func (c Config) ToYaml() []byte {
	// TODO handle error
	yamlstr, _ := yaml.Marshal(c)
	return yamlstr
}

func (c Config) ToYamlFile(p string) {
	// TODO handle error
	ioutil.WriteFile(p, c.ToYaml(), 0600)
}

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
