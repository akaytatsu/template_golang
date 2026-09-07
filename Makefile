SHELL := /bin/bash
ARGS = $(filter-out $@,$(MAKECMDGOALS))
MAKEFLAGS += --silent

.DEFAULT_GOAL := help

# ============================================================================
# Configuração
# ============================================================================
BASE_PATH := $(PWD)
ENV_FILE := src/.env
ENV_SAMPLE := src/.env.sample

# Compose v2 (plugin do Docker). O binário `docker-compose` v1 está EOL.
DOCKER_COMPOSE ?= docker compose
COMPOSE := $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.override.yml
COMPOSE_PROD := $(DOCKER_COMPOSE) -f docker-compose.yml

# Nome base das imagens. Os compose files acrescentam :dev e :prod.
IMAGE_NAME ?= $(notdir $(BASE_PATH))-app
# Tag varrida pelo `security-scan-image` — a de produção, que é a que é publicada.
SCAN_IMAGE ?= $(IMAGE_NAME):prod

# `-include` (e não `include`): o .env é gitignored, então num clone limpo ele
# não existe. Com `include` o make aborta em qualquer target.
-include $(ENV_FILE)
export

# ============================================================================
# Ajuda
# ============================================================================
help: ## Lista os targets disponíveis
	@echo "Uso: make <target> [ARGS]"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Targets marcados com [container] exigem 'make up' antes."

# ============================================================================
# Ambiente
# ============================================================================
_ensure_env:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "Arquivo $(ENV_FILE) não encontrado. Copiando de $(ENV_SAMPLE)..."; \
		cp $(ENV_SAMPLE) $(ENV_FILE); \
		echo "Arquivo $(ENV_FILE) criado com sucesso."; \
	fi

init: _ensure_env ## Prepara o projeto (cria src/.env e instala o gopls)
	go install golang.org/x/tools/gopls@v0.23.0

install_deps: ## Instala as ferramentas de lint/format/hot-reload no host
	@echo "Installing dependencies..."
	go install mvdan.cc/gofumpt@v0.11.0
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
	go install github.com/air-verse/air@v1.67.4

# ============================================================================
# Ciclo de vida dos containers
# ============================================================================
up: _ensure_env ## Sobe todos os serviços
	$(COMPOSE) up -d --remove-orphans

up-prod: _ensure_env ## Sobe apenas o serviço de app com o Dockerfile de produção
	$(COMPOSE_PROD) up -d --remove-orphans

stop: ## Para os serviços
	$(COMPOSE) stop

down: ## Para e remove os containers
	$(COMPOSE) down

restart: ## Reinicia os serviços
	$(COMPOSE) restart

status: ## Mostra o estado dos serviços
	$(COMPOSE) ps

_rebuild: ## Recria as imagens do zero
	$(COMPOSE) down
	$(COMPOSE) build --no-cache --force-rm

sh: ## [container] Abre um shell no serviço (ex.: make sh app)
	$(COMPOSE) exec $(ARGS) bash

run: ## Roda um comando one-off num serviço (ex.: make run app go version)
	$(COMPOSE) run --rm $(ARGS)

# ============================================================================
# Logs
# ============================================================================
log: ## Últimas 200 linhas de log do app
	$(COMPOSE) logs --tail 200 app

logf: ## Segue o log do app
	$(COMPOSE) logs -f --tail 200 app

logs: ## Segue o log de todos os serviços
	$(COMPOSE) logs -f --tail 200

logger: ## Segue o log de um serviço específico (ex.: make logger kafka)
	$(COMPOSE) logs -f --tail 200 $(ARGS)

# ============================================================================
# Dependências Go
# ============================================================================
dep_install: ## [container] Adiciona uma dependência (ex.: make dep_install github.com/foo/bar)
	$(COMPOSE) exec app go get $(ARGS)
	cd src && go get $(ARGS)

auto_install: ## [container] Resolve as dependências faltantes
	$(COMPOSE) exec app go get ./...
	cd src && go get ./...

mod_tidy: ## [container] Roda go mod tidy
	$(COMPOSE) exec app go mod tidy

update-deps: ## Atualiza todas as dependências e roda tidy
	cd src && go get -u ./... && go mod tidy

generate: ## Regenera os mocks (go:generate)
	cd src && go generate ./...

# ============================================================================
# Testes
# ============================================================================
test: ## [container] Roda todos os testes
	$(COMPOSE) exec app gotestsum

test-watch: ## [container] Roda os testes em modo watch
	$(COMPOSE) exec app gotestsum --watch

test-watch-web: ## Sobe a UI do GoConvey em :9090
	go install github.com/smartystreets/goconvey@v1.8.1
	cd src && goconvey -port 9090 -cover

coverage: ## [container] Gera o relatório de cobertura
	$(COMPOSE) exec app go test -coverprofile=coverage.out ./...
	$(COMPOSE) exec app go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: src/coverage.html"

# ============================================================================
# Qualidade de código
# ============================================================================
fmt: ## Formata o código com gofumpt
	gofumpt -w ./src

lint-validate: ## Roda o golangci-lint em modo validação
	cd src/ && golangci-lint run --path-mode=abs --config=".golangci.yml" --timeout=5m

lint-fix: ## Roda o golangci-lint aplicando as correções automáticas
	cd src/ && golangci-lint run --path-mode=abs --config=".golangci.yml" --timeout=5m --fix

lint: fmt lint-fix lint-validate ## Formata, corrige e valida
	@echo "Linting completed successfully."

# ============================================================================
# Docs
# ============================================================================
swagger: ## [container] Regenera a documentação Swagger
	$(COMPOSE) exec app swag init

# ============================================================================
# Utilitários
# ============================================================================
chown_project: ## Devolve a posse dos arquivos ao usuário atual
	sudo chown -R "$(shell id -u):$(shell id -g)" ./

install_generator: ## Instala o gerador de CRUD
	npm install -g generator-go-clean-architecture-crud

update_generator: ## Atualiza o gerador de CRUD
	npm update -g generator-go-clean-architecture-crud

generator_crud: ## Roda o gerador de CRUD
	yo go-clean-architecture-crud

# ============================================================================
# Segurança
# ============================================================================
TRIVY_IMAGE ?= aquasec/trivy:latest
TRIVY_FS = docker run --rm -v $(BASE_PATH):/project -v trivy-cache:/root/.cache/trivy $(TRIVY_IMAGE) fs \
	--scanners vuln,misconfig,secret

vulncheck: ## [container] Checa vulnerabilidades nas dependências Go
	$(COMPOSE) exec app govulncheck ./...

security-scan: ## Varredura de vulnerabilidades (não bloqueante)
	@echo "Executando varredura de vulnerabilidades na pasta . (Trivy Docker)..."
	$(TRIVY_FS) --exit-code 0 --severity HIGH,CRITICAL,MEDIUM /project

security-scan-blocking: ## Varredura bloqueante (exit 1 em HIGH/CRITICAL)
	@echo "Executando varredura de vulnerabilidades (bloqueante)..."
	$(TRIVY_FS) --exit-code 1 --severity HIGH,CRITICAL /project

security-scan-json: ## Varredura com saída em JSON
	@echo "Executando varredura de vulnerabilidades (saída JSON)..."
	$(TRIVY_FS) --exit-code 0 --severity HIGH,CRITICAL,MEDIUM \
		--format json -o /project/trivy-report.json /project

security-scan-table: ## Varredura com saída em tabela
	@echo "Executando varredura de vulnerabilidades (formato tabela)..."
	$(TRIVY_FS) --exit-code 0 --format table /project

security-scan-image: ## Varredura na imagem de produção (SCAN_IMAGE=...)
	@echo "Executando varredura na imagem $(SCAN_IMAGE)..."
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v trivy-cache:/root/.cache/trivy $(TRIVY_IMAGE) image \
		--severity HIGH,CRITICAL \
		--exit-code 1 \
		$(SCAN_IMAGE)

clean-trivy-cache: ## Limpa o cache do Trivy
	@echo "Limpando cache do Trivy..."
	docker volume rm -f trivy-cache

.PHONY: help init install_deps _ensure_env up up-prod stop down restart status \
	_rebuild sh run log logf logs logger dep_install auto_install mod_tidy \
	update-deps generate test test-watch test-watch-web coverage fmt \
	lint-validate lint-fix lint swagger chown_project install_generator \
	update_generator generator_crud vulncheck security-scan \
	security-scan-blocking security-scan-json security-scan-table \
	security-scan-image clean-trivy-cache

# Catch-all: permite `make sh app` sem que o make tente construir o target "app".
# Precisa vir por último.
%:
	@:
