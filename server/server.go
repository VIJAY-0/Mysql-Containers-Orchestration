package server

import (
	"MYSQL-orchestration-API/config"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func addServer(fileName string) ([]byte, error) {

	cmd := exec.Command("docker-compose", "-f", fileName, "up", "-d")
	cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
	output, err := cmd.CombinedOutput()

	return output, err
}

func StartMaster(masterconfigtemplate *config.ServerConfig, Master map[string]*config.ServerConfig, MasterID string) gin.HandlerFunc {

	return func(c *gin.Context) {

		// fmt.Printf("StartMaster ->%s\n", MasterID)

		masterconfig := config.CreateServerConfig(masterconfigtemplate, MasterID)
		Master[MasterID] = masterconfig

		fileName, err := config.MasterCompose(masterconfig, MasterID, "mysql-network")
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

func AddSlave(slaveConfigTemplate *config.ServerConfig, Slaves map[string]*config.ServerConfig) gin.HandlerFunc {

	return func(c *gin.Context) {

		SlaveCntr := len(Slaves)
		SlaveID := fmt.Sprintf("%s%v", slaveConfigTemplate.ContainerName, SlaveCntr)

		// fileName, err := config.SlaveCompose(slaveconfig, SlaveID, "mysql-network")
		slaveconfig := config.CreateServerConfig(slaveConfigTemplate, SlaveID)

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

		Slaves[SlaveID] = slaveconfig
		fmt.Println(*slaveconfig)

		// MysqlServer.initSlave(slaveconfig)
	}
}
