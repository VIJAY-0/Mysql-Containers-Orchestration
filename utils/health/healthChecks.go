package health

import (
	"MYSQL-orchestration-API/config"
	"MYSQL-orchestration-API/server"
	"log"
	"os/exec"
	"strings"
	"time"
)

func HealthCheck(slaves map[string]*config.ServerConfig) {
	for {
		log.Printf("\n\nRunning Health Checks")
		log.Printf("Current Active Slaves :%v\n", len(slaves))
		cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")

		output, _ := cmd.Output()
		outStr := string(output)
		log.Printf("server:%s\n", outStr)

		for SlaveID, serverconfig := range slaves {
			log.Printf("slaveID : %v\n", SlaveID)
			if strings.Contains(outStr, SlaveID) {
				log.Printf("Slave %s is healthy", SlaveID)
				// log.Println(serverconfig)
			} else {
				log.Printf("Slave %s is unhealthy: Restarting...", SlaveID)
				// log.Println(serverconfig)
				filename, err := config.MasterCompose(serverconfig, SlaveID, "mysql-network")
				if err != nil {
					log.Printf("Error occured while restarting :%v", err)
				}
				server.AddServer(filename)
			}
		}
		log.Printf("\n-----------------------------\n\n")
		time.Sleep(500 * time.Millisecond) // Interval between checks
	}
}
