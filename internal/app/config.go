package app

import (
	"os"
	"path/filepath"
)

type Config struct {
	DataDir string
	Addr    string
}

func LoadConfigFromEnv() Config {
	dataDir := os.Getenv("COCKPIT_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	addr := os.Getenv("COCKPIT_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{
		DataDir: dataDir,
		Addr:    addr,
	}
}

func (c Config) DBPath() string {
	return filepath.Join(c.DataDir, "cockpit.db")
}

func (c Config) ArtifactsDir() string {
	return filepath.Join(c.DataDir, "artifacts")
}

