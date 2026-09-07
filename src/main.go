package main

import (
	"app/api"
	"app/config"
	"app/cron"
	"app/infrastructure/postgres"
	"app/kafka"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "time/tzdata" // Required for tzdata to work
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "checa o endpoint /health e sai com 0 (ok) ou 1 (falha)")
	flag.Parse()

	// Usado pelo HEALTHCHECK do Dockerfile: a imagem distroless não tem shell,
	// então o próprio binário faz o probe.
	if *healthcheck {
		if err := api.Healthcheck(); err != nil {
			log.Printf("healthcheck falhou: %v", err)
			os.Exit(1)
		}

		return
	}

	// os.Exit só no main: dentro de run() os defers precisam rodar.
	if err := run(); err != nil {
		log.Printf("erro fatal: %v", err)
		os.Exit(1)
	}
}

func run() error {
	config.ReadEnvironmentVars()

	// Cancelado em SIGINT/SIGTERM para permitir o shutdown ordenado.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cron.StartCronJobs()

	conn := postgres.Connect()
	if err := postgres.Migrations(); err != nil {
		return fmt.Errorf("erro ao rodar as migrations: %w", err)
	}

	// Repository and usecase initialization can be added here if needed
	// usecase := usecase_user.NewService(
	//	repository.NewUserPostgres(conn),
	// )

	defer kafka.Close()

	go kafka.StartKafka(ctx, conn)

	return api.StartWebServer(ctx, conn)
}
