package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type ServerInfo struct {
	ServerName  string `json:"servername"`
	ServerCount int    `json:"servercount"`
}

type ServerConfig struct {
	ContainerName string `yaml:"container_name"`
	RootPassword  string `yaml:"root_password"`
	DatabaseName  string `yaml:"database_name"`
	PortMapping   string `yaml:"port_mapping"`
	Image         string `yaml:"image"`
}

type Config struct {
	MasterServer ServerConfig `yaml:"masterserver"`
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

func StartMaster(masterconfig *ServerConfig) gin.HandlerFunc {

	return func(c *gin.Context) {
		cmd := exec.Command("docker", "run", "-d",
			"--name", masterconfig.ContainerName,
			"-e", fmt.Sprintf("MYSQL_ROOT_PASSWORD=%v", masterconfig.RootPassword),
			"-e", fmt.Sprintf("MYSQL_DATABASE=%s", masterconfig.DatabaseName),
			"-p", masterconfig.PortMapping,
			masterconfig.Image)

		cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")

		output, err := cmd.CombinedOutput()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to start Mysql Master Container",
				"details": string(output),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Sucessfully created container",
			"container": string(output),
		})
	}
}

func main() {

	config, err := LoadConfig()

	if err != nil {
		log.Fatalf("Error loading config :%v", err)
	}

	r := gin.Default()

	// curl http://localhost:8080/init
	r.GET("/createmaster", StartMaster(&config.MasterServer))

	r.Run(":8080")
}
