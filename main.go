package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"redis-like-server/internal/server"
)

func main() {
	// Parse command-line arguments
	port := flag.Int("port", 6379, "Port to listen on")
	maxClients := flag.Int("max-clients", 1000, "Maximum number of concurrent clients")
	readTimeout := flag.Duration("read-timeout", 30*time.Second, "Read timeout for client connections")
	writeTimeout := flag.Duration("write-timeout", 30*time.Second, "Write timeout for client connections")
	flag.Parse()

	// Create server configuration
	config := &server.ServerConfig{
		Port:         *port,
		MaxClients:   *maxClients,
		ReadTimeout:  *readTimeout,
		WriteTimeout: *writeTimeout,
	}

	// Create and start server
	srv := server.NewServer(config)
	
	fmt.Printf("Starting Redis-like server on port %d...\n", config.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}