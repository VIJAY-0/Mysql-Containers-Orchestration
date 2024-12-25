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

type Services struct {
	MasterServer ServerConfig `yaml:"master,omitempty"`
	SlaveServer  ServerConfig `yaml:"slave,omitempty"`
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

func MasterCompose(master ServerConfig, MasterId string, network string) (string, error) {

	var MasterComposed ComposeFile

	MasterComposed.Services = make(map[string]*ServerConfig)
	masterr := master
	MasterComposed.Services[MasterId] = &masterr
	contaierName := fmt.Sprintf("%s-server", MasterId)
	volumeName := fmt.Sprintf("%s-data", MasterId)

	MasterComposed.Services[MasterId].ContainerName = contaierName
	volumeSuff := MasterComposed.Services[MasterId].Volumes[1]

	MasterComposed.Services[MasterId].Volumes[1] = fmt.Sprintf("%s%s", volumeName, volumeSuff)

	MasterComposed.Volumes = make(map[string]map[string]any)
	MasterComposed.Volumes[volumeName] = make(map[string]any)
	MasterComposed.Volumes[volumeName]["name"] = volumeName

	MasterComposed.Networks = make(map[string]map[string]any)
	MasterComposed.Networks[network] = make(map[string]any)
	MasterComposed.Networks[network]["name"] = network

	yamlData, err := yaml.Marshal(&MasterComposed)
	MasterComposed.Services[MasterId].Volumes[1] = volumeSuff
	if err != nil {
		fmt.Printf("Error marshalling to YAML: %v\n", err)
		return "", err
	}

	// fileName := fmt.Sprintf("%s.yaml", MasterId)
	fileName := fmt.Sprintf("DockerCompose\\docker-compose%s.yaml", MasterId)

	err = os.WriteFile(fileName, yamlData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return "", err
	}

	fmt.Printf("YAML file '%s' created successfully!\n", fileName)

	return fileName, nil

}

// func SlaveCompose(master ServerConfig, MasterId string, network string) (string, error) {
// 	var MasterComposed ComposeFile

// 	MasterComposed.Services.SlaveServer = master

// 	contaierName := fmt.Sprintf("%s-server", MasterId)
// 	volumeName := fmt.Sprintf("%s-data", MasterId)

// 	MasterComposed.Services.SlaveServer.ContainerName = contaierName
// 	volumeSuff := MasterComposed.Services.SlaveServer.Volumes[1]
// 	MasterComposed.Services.SlaveServer.Volumes[1] = fmt.Sprintf("%s%s", volumeName, volumeSuff)

// 	MasterComposed.Volumes = make(map[string]map[string]any)
// 	MasterComposed.Volumes[volumeName] = make(map[string]any)
// 	MasterComposed.Volumes[volumeName]["name"] = volumeName

// 	MasterComposed.Networks = make(map[string]map[string]any)
// 	MasterComposed.Networks[network] = make(map[string]any)
// 	MasterComposed.Networks[network]["name"] = network

// 	yamlData, err := yaml.Marshal(&MasterComposed)
// 	MasterComposed.Services.SlaveServer.Volumes[1] = volumeSuff
// 	if err != nil {
// 		fmt.Printf("Error marshalling to YAML: %v\n", err)
// 		return "", err
// 	}

// 	fileName := fmt.Sprintf("DockerCompose\\docker-compose%s.yaml", MasterId)
// 	err = os.WriteFile(fileName, yamlData, 0644)
// 	if err != nil {
// 		fmt.Printf("Error writing to file: %v\n", err)
// 		return "", err
// 	}

// 	fmt.Printf("YAML file '%s' created successfully!\n", fileName)
// 	return fileName, nil
// }
