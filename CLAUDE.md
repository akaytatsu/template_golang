# CLAUDE.md — Go Clean Architecture Template

## Project Overview

This is a Go template project implementing **Clean Architecture** principles for RESTful APIs. It provides a solid foundation with JWT authentication, PostgreSQL (GORM), Kafka messaging, Swagger docs, and a comprehensive test structure.

## Go Version

- **Required:** Go 1.26.5+
- **Docker image:** `golang:1.26.5-bookworm`

## Key Commands

### Development
```bash
make init          # Copy .env.sample to .env
make up            # Start all services (app, db, kafka, mail)
make stop          # Stop all services
make logs          # Tail all service logs
make log           # Tail app logs only
make sh            # Open shell in app container (e.g., make sh app)
```

### Testing
```bash
make test          # Run all tests
make coverage      # Run tests with coverage report
make test-watch    # Watch mode tests
make test-watch-web # GoConvey web UI on :9090
```

### Code Quality
```bash
make fmt           # Format code with gofumpt
make lint          # Full lint (fmt + fix + validate)
make lint-validate # golangci-lint check only
```

### Dependencies
```bash
make dep_install <pkg>  # Add dependency (in container + local)
make auto_install       # go get ./... in container + local
make mod_tidy           # go mod tidy in container
make update-deps        # go get -u ./... + tidy
```

### Security
```bash
make security-scan           # Trivy filesystem scan
make security-scan-blocking  # Trivy blocking scan (exit 1 on HIGH/CRITICAL)
make security-scan-image     # Trivy image scan
make vulncheck               # Go vulnerability check via govulncheck
```

### Docs
```bash
make swagger    # Generate Swagger/OpenAPI docs
```

## Architecture

```
src/
├── main.go                   # Entry point: config → db → kafka → webserver
├── api/                      # HTTP layer
│   ├── api.go                # Router setup, middleware registration
│   ├── handlers/             # HTTP handlers
│   └── middleware/           # Auth & custom middleware
├── entity/                   # Domain entities (GORM models)
├── usecase/                  # Business logic (interfaces + services)
├── infrastructure/
│   ├── postgres/             # DB connection & migrations
│   └── repository/           # Repository implementations
├── kafka/                    # Kafka consumer/producer setup
├── config/                   # Environment variable parsing
│   ├── environment.go        # env loader
│   └── model.go              # Config struct
├── cron/                     # Scheduled jobs (gocron)
├── pkg/
│   ├── logger/               # Structured logging
│   ├── utils/                # General utilities
│   └── testing_utils/        # Test helpers
├── mocks/                    # Generated mocks (go.uber.org/mock)
├── certs/                    # TLS certificates
└── docs/                     # Swagger generated docs
```

## Services (docker-compose)

| Service | Image | Port |
|---------|-------|------|
| app | Local build | 8080 (API), 2345 (Delve) |
| db | postgres:16-bookworm | 5432 |
| kafka | confluentinc/cp-kafka:8.2.2 | 9092 |
| zookeeper | confluentinc/cp-zookeeper:8.2.2 | 2181 |
| webkafka | provectuslabs/kafka-ui:v0.7.2 | 9030 |
| mail | mailhog/mailhog:v1.1.0 | 8025 |

## Development Tools (installed in Dockerfile-dev)

| Tool | Version | Purpose |
|------|---------|---------|
| air | v1.67.1 | Hot reload |
| dlv | v1.27.0 | Debugger |
| gotestsum | v1.13.0 | Test runner |
| mockgen | v0.6.0 | Mock generation |
| goconvey | v1.8.1 | BDD test framework |
| swag | v1.16.6 | Swagger doc generation |
| govulncheck | v1.6.0 | Vulnerability scanning |

## Code Style

- Formatter: `gofumpt` (stricter gofmt)
- Imports: `goimports`
- Linter: `golangci-lint` v2 (config: `src/.golangci.yml`)
- Linters enabled: bodyclose, gocritic, gosec, misspell, noctx, nolintlint, rowserrcheck, sqlclosecheck, staticcheck, tparallel, whitespace

## Environment Variables

Key env vars in `src/.env`:
- `LOG_LEVEL`, `GIN_MODE`, `GORM_LOG_LEVEL` — logging control
- `POSTGRES_*` — database connection
- `KAFKA_*` — Kafka bootstrap server and client config
- `JWT_SECRET_KEY` — JWT signing key
- `EMAIL_*` — SMTP/mail config
- `DEFAULT_ADMIN_*` — initial admin credentials

## Key Dependencies

| Package | Purpose |
|---------|---------|
| gin-gonic/gin | HTTP framework |
| gorm.io/gorm + driver/postgres | ORM |
| confluentinc/confluent-kafka-go/v2 | Kafka client (CGO) |
| golang-jwt/jwt/v5 | JWT auth |
| go-co-op/gocron/v2 | Cron scheduling |
| swaggo/swag + gin-swagger | API docs |
| stretchr/testify | Test assertions |
| smartystreets/goconvey | BDD testing |
| go.uber.org/mock | Mock generation |
| golang.org/x/crypto | bcrypt password hashing |

## Build Notes

- **CGO is required** (librdkafka dependency)
- Production image: multi-stage build → `gcr.io/distroless/base-debian12:nonroot`
- Binary stripped and statically linked
- Build flags: `-ldflags="-w -s -linkmode external -extldflags '-static -Wl,-z,relro,-z,now'"` 
- `GOWORK=off` during Docker builds
