package server

import (
	"MYSQL-orchestration-API/config"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func dockerContainerCommand(serverconfig *config.ServerConfig) []string {
	args := []string{"run", "-d"}
	if serverconfig.ContainerName != "" {
		args = append(args, "--name", serverconfig.ContainerName)
	}
	if serverconfig.RootPassword != "" {
		args = append(args, "-e", fmt.Sprintf("MYSQL_ROOT_PASSWORD=%v", serverconfig.RootPassword))
	} else {
		args = append(args, "-e", "MYSQL_ROOT_PASSWORD=root")
	}

	if serverconfig.DatabaseName != "" {
		args = append(args, "-e", fmt.Sprintf("MYSQL_DATABASE=%s", serverconfig.DatabaseName))
	} else {
		args = append(args, "-e", "MYSQL_DATABASE=mydb")
	}

	if serverconfig.PortMapping != "" {
		args = append(args, "-p", serverconfig.PortMapping)
	}
	args = append(args, serverconfig.Image)

	return args
}

func addMYSQLServer(serverconfig *config.ServerConfig) ([]byte, error) {
	args := dockerContainerCommand(serverconfig)

	cmd := exec.Command("docker", args...)

	cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
	output, err := cmd.CombinedOutput()

	return output, err
}

func StartMaster(masterconfig *config.ServerConfig) gin.HandlerFunc {

	return func(c *gin.Context) {

		output, err := addMYSQLServer(masterconfig)

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
				"response": "Failed to add Mysql Slave Container",
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

func AddSlave(slaveconfig *config.ServerConfig, Slaves map[string]string) gin.HandlerFunc {

	return func(c *gin.Context) {

		ContainerName := slaveconfig.ContainerName
		SlaveCntr := len(Slaves)
		NewContainerName := fmt.Sprintf("%s%v", slaveconfig.ContainerName, SlaveCntr)
		slaveconfig.ContainerName = NewContainerName
		output, err := addMYSQLServer(slaveconfig)
		slaveconfig.ContainerName = ContainerName

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
		Slaves[NewContainerName] = NewContainerName
	}
}
