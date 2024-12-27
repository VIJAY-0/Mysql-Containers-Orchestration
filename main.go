package main

import (
	"MYSQL-orchestration-API/config"
	"MYSQL-orchestration-API/server"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
)

func cleanup(Slaves map[string]*config.ServerConfig, Master map[string]*config.ServerConfig) gin.HandlerFunc {

	for key, serverconfig := range Master {
		Slaves[key] = serverconfig
		delete(Master, key)
	}

	return func(c *gin.Context) {
		for key, serverconfig := range Slaves {
			// value := Slaves[key]
			container := fmt.Sprintf("%s-server", key)
			cmd := exec.Command("docker", "rm", "-f", container)
			cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
			output, err := cmd.CombinedOutput()
			fmt.Printf("Cleaned %s", string(output))
			if err != nil {
				// return
			}
			volume := fmt.Sprintf("%s-data", key)
			cmd = exec.Command("docker", "volume", "rm", volume)
			cmd.Env = append(cmd.Env, "DOCKER_HOST=npipe:////./pipe/docker_engine")
			output, err = cmd.CombinedOutput()

			if err != nil {
				// return
			}
			fmt.Printf("Cleaned %s", string(output))
			delete(Slaves, key)

			fmt.Printf("Cleaned %s", serverconfig)
			// free(serverconfig)
		}

	}
}

func main() {

	compose, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Error loading config :%v", err)
	}

	r := gin.Default()

	Master := make(map[string]*config.ServerConfig)
	Slaves := make(map[string]*config.ServerConfig)

	MasterID := "master"
	Master[MasterID] = nil

	// // curl http://localhost:8080/createmaster
	r.GET("/createmaster", server.StartMaster(compose.Services["master"], Master, MasterID))

	// curl http://localhost:8080/addslave

	r.GET("/addslave", server.AddSlave(compose.Services["slave"], Slaves))

	r.GET("/cleanup", cleanup(Slaves, Master))

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %s\n", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) // Capture Ctrl+C
	<-quit                            // Block until signal is received

	log.Println("Running CleanUp...")
	cleanup(Slaves, Master)
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s\n", err)
	}

	log.Println("Server exiting")

}
