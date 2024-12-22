package main

import (
	"MYSQL-orchestration-API/config"
	"MYSQL-orchestration-API/server"
	"fmt"
	"log"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func cleanup(Slaves map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for key := range Slaves {
			// value := Slaves[key]
			cmd := exec.Command("docker", "rm", "-f", key)
			cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return
			}

			fmt.Printf("Cleaned %s", string(output))
			delete(Slaves, key)
		}

	}
}

func main() {

	config, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Error loading config :%v", err)
	}

	r := gin.Default()

	// curl http://localhost:8080/createmaster
	r.GET("/createmaster", server.StartMaster(&config.MasterServer))

	// curl http://localhost:8080/addslave
	Slaves := make(map[string]string)

	r.GET("/addslave", server.AddSlave(&config.SlaveServer, Slaves))

	r.GET("/cleanup", cleanup(Slaves))
	r.Run(":8080")
}
