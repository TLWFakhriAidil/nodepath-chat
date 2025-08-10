FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git build-base ca-certificates

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/server ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

# Create app directory
RUN mkdir -p /app

# Copy binary from builder
COPY --from=builder /app/bin/server /app/server

# Copy templates and static files
COPY --from=builder /src/templates /app/templates
COPY --from=builder /src/static /app/static

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Set default environment variables
ENV PORT=8080
ENV APP_ENV=production

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/healthz || exit 1

# Run the application
CMD ["/app/server"]