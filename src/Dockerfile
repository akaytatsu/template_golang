# ============================================================================
# STAGE 1: Builder
# ============================================================================
FROM golang:1.24-bookworm AS builder

# Environment variables for build
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct

# Install build dependencies only (no cache to reduce layer size)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    gcc \
    g++ \
    libc6-dev \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Set library paths for CGO
ENV CGO_CFLAGS="-O2 -g" \
    CGO_LDFLAGS="-lresolv" \
    PKG_CONFIG_PATH="/usr/lib/x86_64-linux-gnu/pkgconfig"

WORKDIR /build

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Install swag and generate Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    /go/bin/swag init

# Build the application with security flags
RUN go build -ldflags="-w -s \
    -linkmode external \
    -extldflags '-static -Wl,-z,relro,-z,now'" \
    -a -installsuffix cgo \
    -o /build/app_bin .

# Strip debug symbols for smaller binary
RUN strip --strip-unneeded /build/app_bin 2>/dev/null || true

# Create tmp directory in builder stage (will be copied to final stage)
RUN mkdir -p /tmp && chmod 1777 /tmp

# ============================================================================
# STAGE 2: Final Runtime Image (Distroless)
# ============================================================================
FROM gcr.io/distroless/base-debian12:nonroot

# Security: Run as non-root user (distroless nonroot = UID 65532)
USER nonroot:nonroot

# Set working directory
WORKDIR /app

# Copy only necessary files from builder
COPY --from=builder --chown=nonroot:nonroot /build/app_bin /app/app_bin
COPY --from=builder --chown=nonroot:nonroot /build/docs /app/docs

# Copy CA certificates for HTTPS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy tmp directory from builder (distroless images have no shell/commands)
COPY --from=builder --chown=nonroot:nonroot /tmp /tmp

# Expose application port
EXPOSE 8080

# Use ENTRYPOINT for fixed command, CMD for default arguments
ENTRYPOINT ["/app/app_bin"]