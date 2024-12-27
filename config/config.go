package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Environment struct {
	MYSQL_ROOT_PASSWORD string `yaml:"MYSQL_ROOT_PASSWORD,omitempty"`
	MYSQL_DATABASE      string `yaml:"MYSQL_DATABASE,omitempty"`
	MYSQL_USER          string `yaml:"MYSQL_USER,omitempty"`
	MYSQL_PASSWORD      string `yaml:"MYSQL_PASSWORD,omitempty"`
}

func (env *Environment) DeepCopy() *Environment {
	if env == nil {
		return nil
	}
	return &Environment{
		MYSQL_ROOT_PASSWORD: env.MYSQL_ROOT_PASSWORD,
		MYSQL_DATABASE:      env.MYSQL_DATABASE,
		MYSQL_USER:          env.MYSQL_USER,
		MYSQL_PASSWORD:      env.MYSQL_PASSWORD,
	}
}

func (config *ServerConfig) DeepCopy() *ServerConfig {
	if config == nil {
		return nil
	}
	return &ServerConfig{
		Image:         config.Image,
		ContainerName: config.ContainerName,
		Restart:       config.Restart,
		Environment:   *config.Environment.DeepCopy(),            // Deep copy of the Environment
		Ports:         append([]string(nil), config.Ports...),    // Copy the slice
		Volumes:       append([]string(nil), config.Volumes...),  // Copy the slice
		Networks:      append([]string(nil), config.Networks...), // Copy the slice
		Entrypoint:    config.Entrypoint,
	}
}

type ServerConfig struct {
	Image         string      `yaml:"image,omitempty"`
	ContainerName string      `yaml:"container_name,omitempty"`
	Restart       string      `yaml:"restart,omitempty"`
	Environment   Environment `yaml:"environment,omitempty"`
	Ports         []string    `yaml:"ports,omitempty"`
	Volumes       []string    `yaml:"volumes,omitempty"`
	Networks      []string    `yaml:"networks,omitempty"`
	Entrypoint    string      `yaml:"entrypoint,omitempty"`
}

type ComposeFile struct {
	Version  string                    `yaml:"version,omitempty"`
	Services map[string]*ServerConfig  `yaml:"services,omitempty"`
	Volumes  map[string]map[string]any `yaml:"volumes,omitempty"`
	Networks map[string]map[string]any `yaml:"networks,omitempty"`
}

func LoadConfig() (*ComposeFile, error) {

	data, err := os.ReadFile(".\\config\\SERVER_CONFIG.yaml")

	if err != nil {
		return nil, fmt.Errorf("error reading config file :%v", err)
	}
	var compose ComposeFile
	err = yaml.Unmarshal(data, &compose)

	if err != nil {
		return nil, fmt.Errorf("error parsing yaml :%v", err)
	}
	return &compose, nil
}

func CreateServerConfig(template *ServerConfig, ServerId string) *ServerConfig {

	serverconfig := template.DeepCopy()

	contaierName := fmt.Sprintf("%s-server", ServerId)
	volumeName := fmt.Sprintf("%s-data", ServerId)

	serverconfig.ContainerName = contaierName
	volumeSuff := template.Volumes[1]
	serverconfig.Volumes[1] = fmt.Sprintf("%s%s", volumeName, volumeSuff)
	return serverconfig
}

func MasterCompose(serverconfig *ServerConfig, ServerId string, network string) (string, error) {

	var MasterComposed ComposeFile

	MasterComposed.Services = make(map[string]*ServerConfig)
	volumeName := fmt.Sprintf("%s-data", ServerId)

	MasterComposed.Services[ServerId] = serverconfig
	MasterComposed.Volumes = make(map[string]map[string]any)
	MasterComposed.Volumes[volumeName] = make(map[string]any)
	MasterComposed.Volumes[volumeName]["name"] = volumeName

	MasterComposed.Networks = make(map[string]map[string]any)
	MasterComposed.Networks[network] = make(map[string]any)
	MasterComposed.Networks[network]["name"] = network

	yamlData, err := yaml.Marshal(&MasterComposed)

	if err != nil {
		fmt.Printf("Error marshalling to YAML: %v\n", err)
		return "", err
	}

	// fileName := fmt.Sprintf("%s.yaml", MasterId)
	fileName := fmt.Sprintf("DockerCompose\\docker-compose%s.yaml", ServerId)

	err = os.WriteFile(fileName, yamlData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return "", err
	}

	fmt.Printf("YAML file '%s' created successfully!\n", fileName)

	return fileName, nil

}
