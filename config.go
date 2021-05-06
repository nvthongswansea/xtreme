package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config holds configuration of xtreme.
type Config struct {
	LogLevel string         `yaml:"log_level"`
	Backend  BackendConfig  `yaml:"backend"`
	Frontend FrontendConfig `yaml:"frontend"`
}

// BackendConfig holds properties of backend's configuration.
type BackendConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	UploadDir string `yaml:"upload_dir"`
}

// FrontendConfig holds properties of frontend's configuration.
type FrontendConfig struct {
}

// NewConfig returns a new Config.
func NewConfig(configFile string) (*Config, error) {
	config := &Config{}
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
