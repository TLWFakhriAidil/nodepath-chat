# Build stage
FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy entire project
COPY . .

# Change to src directory where go.mod is located
WORKDIR /app/src

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 go build -o /app/whatsapp .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/whatsapp .

# Copy templates and static files
COPY --from=builder /app/src/views ./views
COPY --from=builder /app/src/statics ./statics

# Create entrypoint script directly
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'if [ -n "$DATABASE_URL" ]; then' >> /app/entrypoint.sh && \
    echo '    export DB_URI="$DATABASE_URL"' >> /app/entrypoint.sh && \
    echo '    echo "DB_URI set from DATABASE_URL"' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo 'echo "Starting WhatsApp Multi-Device in REST mode..."' >> /app/entrypoint.sh && \
    echo 'exec /app/whatsapp rest' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Create directories for storage
RUN mkdir -p /app/storages /app/sessions

# Expose port
EXPOSE 3000

# Set entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]