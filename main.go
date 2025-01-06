package main

import (
	"MYSQL-orchestration-API/LoadBalancer"
	"MYSQL-orchestration-API/config"
	"MYSQL-orchestration-API/server"
	"MYSQL-orchestration-API/utils/health"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
)

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

	var MasterDSN string
	// // curl http://localhost:8080/createmaster
	r.GET("/createmaster", server.StartMaster(compose.Services["master"], Master, MasterID, &MasterDSN))
	r.GET("/initmaster", server.InitMaster(Master, MasterID))
	r.GET("/master", server.PrintMaster(Master, MasterID))

	// curl http://localhost:8080/addslave

	var SlaveDSNs []string

	r.GET("/addslave", server.AddSlave(compose.Services["slave"], Slaves, Master, MasterID, &SlaveDSNs))
	r.GET("/removeslave", server.RemoveSlave(compose.Services["slave"], Slaves))
	r.GET("/listslaves", server.ListSlave(&SlaveDSNs))
	r.GET("/cleanup", server.Cleanup(Slaves, Master))

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

	//LOAD BALANCER
	PORT := "3305"

	lb, err := LoadBalancer.NewLoadBalancer(SlaveDSNs)
	if err != nil {
		log.Fatalf("Failed to initialize load balancer: %v", err)
	}

	lb.SlaveDSNs = &SlaveDSNs
	lb.MasterDSN = &MasterDSN

	listener, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Fatalf("Failed to start Proxy server: %v", err)
	}
	defer listener.Close()

	log.Printf("LoadBalancer Proxy listening on Port:,%s", PORT)

	go health.HealthCheck(Slaves)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go lb.HandleConnection(clientConn)
	}

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) // Capture Ctrl+C
	<-quit                            // Block until signal is received

	log.Println("Running CleanUp...")
	// cleanup.Cleanup(Slaves, Master)

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s\n", err)
	}

	log.Println("Server exiting")

}
