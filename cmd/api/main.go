package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

 
	"goat/internal/database"
	"goat/internal/engine"
	"goat/internal/server"
	"goat/plugins/slack"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}




func main() {
 
	// 1. Initialize Database
	// NOTE: Your database.New() is hardcoded to PostgreSQL. You might want to
	// make this configurable based on environment variables.
	dbService := database.New()
 
	// 2. Initialize Plugin Registry and register plugins
	registry := engine.NewPluginRegistry()
	registry.Register("slack.send_message", slack.NewSendMessageExecutor())
	// ... register other plugins here
 
	// 3. Initialize the Workflow Engine
	workflowEngine := engine.New(dbService, registry)
 
	// 4. Initialize the HTTP Server
	// NOTE: You will need to modify `server.NewServer()` to accept the engine
	// so your API handlers can use it to trigger workflows.
	apiServer := server.NewServer(workflowEngine)

	server := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)
	go gracefulShutdown(apiServer, done)

	err := server.ListenAndServe()
	err := apiServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
