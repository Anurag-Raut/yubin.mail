package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
)

type client_config struct {
	ServerName   string `yaml:"ServerName"`
	Organization string `yaml:"Organization"`
	IP           string `yaml:"IP"`
}

func loadConfig() client_config {
	_, currFilename, _, _ := runtime.Caller(0)

	dir := filepath.Dir(currFilename)
	configFilePath := filepath.Join(dir, "config.yaml")
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		panic(err)
	}
	var cfg client_config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

var ClientConfig = loadConfig()
