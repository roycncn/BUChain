package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	LogLevel   string `yaml:"log_level"`
	LogOutput  string `yaml:"log_output"`
	PeersList  string `yaml:"peers_list"`
	NodePort   string `yaml:"node_port"`
	RestPort   string `yaml:"rest_port"`
	PrivateKey string `yaml:"private_key"`
}

// Setup init config
func New(path string) (*Config, error) {
	// config global config instance
	var config = new(Config)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
