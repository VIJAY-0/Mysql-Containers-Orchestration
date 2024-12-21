package main

import (
	"MYSQL-orchestration-API/config"
	"MYSQL-orchestration-API/server"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	config, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Error loading config :%v", err)
	}

	r := gin.Default()

	// curl http://localhost:8080/init
	r.GET("/createmaster", server.StartMaster(&config.MasterServer))

	r.Run(":8080")
}
