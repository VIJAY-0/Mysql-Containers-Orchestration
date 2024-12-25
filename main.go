package main

import (
	"MYSQL-orchestration-API/config"
	"MYSQL-orchestration-API/server"
	"fmt"
	"log"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func cleanup(Slaves map[string]string, Master string) gin.HandlerFunc {
	Slaves[Master] = Master
	return func(c *gin.Context) {
		for key := range Slaves {
			// value := Slaves[key]
			container := fmt.Sprintf("%s-server", key)
			cmd := exec.Command("docker", "rm", "-f", container)
			cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
			output, err := cmd.CombinedOutput()
			if err != nil {
				// return
			}

			volume := fmt.Sprintf("%s-data", key)
			cmd = exec.Command("docker", "rm", "-f", volume)
			cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
			output, err = cmd.CombinedOutput()

			if err != nil {
				// return
			}
			fmt.Printf("Cleaned %s", string(output))
			delete(Slaves, key)
		}

	}
}

func main() {

	compose, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Error loading config :%v", err)
	}

	r := gin.Default()

	Master := "master"
	Slaves := make(map[string]string)

	// // curl http://localhost:8080/createmaster
	r.GET("/createmaster", server.StartMaster(*compose.Services["master"], Master))

	// curl http://localhost:8080/addslave

	r.GET("/addslave", server.AddSlave(*compose.Services["slave"], Slaves))

	r.GET("/cleanup", cleanup(Slaves, Master))
	r.Run(":8080")
}
