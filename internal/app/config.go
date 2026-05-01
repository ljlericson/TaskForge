package app

import (
	"os"

	"github.com/ljlericson/TaskForge/internal/registry"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig            `yaml:"server"`
	Logging LoggingConfig           `yaml:"logging"`
	Session SessionConfig           `yaml:"session"`
	Workers []registry.WorkerConfig `yaml:"workers"`
}

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Timeout int    `yaml:"timeout"`
}

type LoggingConfig struct {
	Path string `yaml:"path"`
}

type SessionConfig struct {
	Key string `yaml:"key"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
