package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	ContainerName string `yaml:"container_name"`
	RootPassword  string `yaml:"root_password"`
	DatabaseName  string `yaml:"database_name"`
	PortMapping   string `yaml:"port_mapping"`
	Image         string `yaml:"image"`
}

type Config struct {
	MasterServer ServerConfig `yaml:"masterserver"`
	SlaveServer  ServerConfig `yaml:"slaveserver"`
}

func LoadConfig() (*Config, error) {

	data, err := os.ReadFile("config.yaml")

	if err != nil {
		return nil, fmt.Errorf("error reading config file :%v", err)
	}
	var config Config
	err = yaml.Unmarshal(data, &config)

	if err != nil {
		return nil, fmt.Errorf("error parsing yaml :%v", err)
	}
	return &config, nil
}
