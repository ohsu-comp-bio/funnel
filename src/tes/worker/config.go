package tesTaskEngineWorker

// Worker configuration.
type Config struct {
  ID          string
	MasterAddr  string
	WorkDir     string
	Timeout     int
	NumWorkers  int
	AllowedDirs []string
	LogPath     string
}

// NewConfig returns a new worker config instance with default values.
func NewConfig() Config {
	return Config{
		MasterAddr: "localhost:9090",
		WorkDir:    "volumes",
		Timeout:    -1,
		NumWorkers: 4,
	}
}
