package server

import (
	"fmt"
	"net/http"
	"os/exec"

	"MYSQL-orchestration-API/config"

	"github.com/gin-gonic/gin"
)

func StartMaster(masterconfig *config.ServerConfig) gin.HandlerFunc {

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
