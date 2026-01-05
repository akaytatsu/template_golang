package main

import (
	"app/api"
	"app/config"
	"app/cron"
	"app/infrastructure/postgres"
	"app/kafka"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "time/tzdata" // Required for tzdata to work
)

func main() {
	if err := config.ReadEnvironmentVars(); err != nil {
		log.Fatalf("Failed to read environment variables: %v", err)
	}

	cron.StartCronJobs()

	postgres.Connect()
	postgres.Migrations()

	// Start Kafka consumer in background
	go kafka.StartKafka()

	// Start web server (handles graceful shutdown internally)
	api.StartWebServer()

	// Cleanup after server shutdown
	log.Println("Application shutting down...")
	// TODO: Add cleanup for database connections, Kafka consumer, etc.
}
