package main

import (
	"app/api"
	"app/cron"
	"app/entity"
	"app/infrastructure/postgres"
)

func main() {
	cron.StartCronJobs()

	db, err := postgres.Connect()

	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&entity.EntityUser{})

	api.StartWebServer()
}
