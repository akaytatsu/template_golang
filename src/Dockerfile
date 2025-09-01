FROM golang:1.24-trixie
ARG PROJECT_NAME="app"
ARG VERSION="dev"
ARG COMMIT="none"
ENV CGO_ENABLED=1
ENV GOPROXY=https://proxy.golang.org

WORKDIR /app

COPY . .

RUN apt update -y && apt install -y gcc libc-dev librdkafka-dev
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/app_bin .

RUN mkdir -p /app/tmp

EXPOSE 8080

ENTRYPOINT ["/app/app_bin"]