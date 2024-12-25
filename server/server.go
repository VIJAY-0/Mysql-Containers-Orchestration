package server

import (
	"MYSQL-orchestration-API/config"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func getARGS(serverconfig config.ServerConfig) []string {
	args := []string{"run", "-d"}
	if serverconfig.ContainerName != "" {
		args = append(args, "--name", serverconfig.ContainerName)
	}

	if serverconfig.Restart == "always" {
		args = append(args, "--restart", "always")
	}

	if serverconfig.Environment.MYSQL_ROOT_PASSWORD != "" {
		args = append(args, "-e", fmt.Sprintf("MYSQL_ROOT_PASSWORD=%v", serverconfig.Environment.MYSQL_ROOT_PASSWORD))
	} else {
		args = append(args, "-e", "MYSQL_ROOT_PASSWORD=root")
	}

	if serverconfig.Environment.MYSQL_DATABASE != "" {
		args = append(args, "-e", fmt.Sprintf("MYSQL_DATABASE=%s", serverconfig.Environment.MYSQL_DATABASE))
	} else {
		args = append(args, "-e", "MYSQL_DATABASE=mydb")
	}

	args = append(args, "-v", serverconfig.Volumes[0])
	args = append(args, "-v", serverconfig.Volumes[1])

	for _, portMapping := range serverconfig.Ports {
		args = append(args, "-p", portMapping)
	}

	for _, network := range serverconfig.Networks {
		args = append(args, "--network", network)
	}

	args = append(args, serverconfig.Image)

	args = append(args, serverconfig.Entrypoint)

	return args
}

func addServer(fileName string) ([]byte, error) {

	cmd := exec.Command("docker-compose", "-f", fileName, "up", "-d")
	cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
	output, err := cmd.CombinedOutput()

	return output, err
}

func StartMaster(masterconfig config.ServerConfig, Master string) gin.HandlerFunc {

	return func(c *gin.Context) {

		fileName, err := config.MasterCompose(masterconfig, Master, "mysql-network")
		output, err := addServer(fileName)

		if err != nil {

			outputstr := string(output)
			if len(outputstr) > 0 && strings.Contains(outputstr, "is already in use by container") {
				c.JSON(http.StatusInternalServerError, gin.H{
					"response": "Container Already running by same name",
					"error":    fmt.Sprintf(" %v", err),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"response": "Failed to add Mysql Master Container",
				"error":    fmt.Sprintf(" %v", err),
				"details":  string(output),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Sucessfully created container",
			"container": string(output),
		})
	}
}

func AddSlave(slaveconfig config.ServerConfig, Slaves map[string]string) gin.HandlerFunc {

	return func(c *gin.Context) {

		SlaveCntr := len(Slaves)
		SlaveID := fmt.Sprintf("%s%v", slaveconfig.ContainerName, SlaveCntr)

		// fileName, err := config.SlaveCompose(slaveconfig, SlaveID, "mysql-network")
		fileName, err := config.MasterCompose(slaveconfig, SlaveID, "mysql-network")
		output, err := addServer(fileName)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"response": "Failed to add Mysql Slave Container",
				"error":    fmt.Sprintf(" %v", err),
				"details":  string(output),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Sucessfully Added Slave container",
			"container": string(output),
		})
		Slaves[SlaveID] = SlaveID
	}
}
