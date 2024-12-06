# Template Golang

### Overview
Template para projeto golang
### Configuração
- Instalar o golang: [Guia de instalação](https://go.dev/doc/install)
- Build da aplicação: `$ make init`
### Ambiente
- Portas
    - Aplicação: `localhost:8080`
    - Interface Kafka: `localhost:9030`
### Padrões de Desenvolvimento
- [Clean Architecture](https://medium.com/luizalabs/descomplicando-a-clean-architecture-cf4dfc4a1ac6).
### Comandos
- Iniciar Aplicação: `$ make up`
- Reiniciar Aplicação: `$ make restart`
- Atualizar/Gerar mocks: `$ make generate`
- Testes: `$ make test` ou `$ make test-watch`