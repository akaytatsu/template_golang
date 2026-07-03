package cron

import (
	"log"
	"time"

	gocron "github.com/go-co-op/gocron/v2"
)

func StartCronJobs() {
	s, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		log.Printf("Error creating gocron scheduler: %v", err)
		return
	}

	s.Start()
}
