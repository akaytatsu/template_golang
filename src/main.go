package main

import (
	"app/api"
	"app/config"
	"app/cron"
	"app/infrastructure/postgres"
	"app/kafka"

	_ "time/tzdata" // Required for tzdata to work
)

func main() {
	config.ReadEnvironmentVars()

	cron.StartCronJobs()

	postgres.Connect()
	postgres.Migrations()

	// Repository and usecase initialization can be added here if needed
	// conn := postgres.Connect()
	// usecase := usecase_user.NewService(
	//	repository.NewUserPostgres(conn),
	// )

	go kafka.StartKafka()

	api.StartWebServer()
}
