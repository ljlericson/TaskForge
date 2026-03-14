package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/ljlericson/TaskForge/internal/api"
	"github.com/ljlericson/TaskForge/internal/console"
	"gopkg.in/yaml.v3"
)

type config struct {
	Server  serverConfig  `yaml:"server"`
	Logging loggingConfig `yaml:"logging"`
	Session sessionConfig `yaml:"session"`
}

type serverConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Timeout int    `yaml:"timeout"`
}

type loggingConfig struct {
	Path string `yaml:"path"`
}

type sessionConfig struct {
	Key string `yaml:"key"`
}

func main() {
	config, err := loadConfig("config/server.yml")
	if err != nil {
		panic(err)
	}

	logDir := filepath.Dir(config.Logging.Path)
	err3 := os.MkdirAll(logDir, os.ModePerm)
	if err3 != nil {
		panic(err3)
	}

	logFile, err2 := os.Create(config.Logging.Path)
	if err2 != nil {
		panic(err2)
	}

	c := console.New(logFile)
	go server(config.Server.Host, c)

	go func() {
		for cmd := range c.Input() {
			switch cmd {
			case "stop":
				c.Stop()
				return
			case "logo":
				c.Log(api.LogoStr)
			default:
				c.Log("Command " + cmd + " is not a valid command")
			}
		}
	}()

	if err := c.Run(); err != nil {
		panic(err)
	}
}

func server(host string, c *console.Console) {
	r := chi.NewRouter()
	r.Use(console.RequestLogger(c))
	api.ConfigureRoutes(r)
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		panic(err)
	}
}

func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
