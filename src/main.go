package main

import (
	"app/api"
	"app/cron"
)

func main() {
	cron.StartCronJobs()

	api.StartWebServer()
}
