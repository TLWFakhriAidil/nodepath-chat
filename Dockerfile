# Frontend build stage
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files and npm configuration
COPY package*.json .npmrc ./

# Install dependencies
RUN npm ci

# Copy source code and configuration files needed for build
COPY src/ ./src/
COPY public/ ./public/
COPY index.html ./
COPY vite.config.ts ./
COPY tsconfig*.json ./
COPY tailwind.config.ts ./
COPY postcss.config.js ./
COPY components.json ./
COPY eslint.config.js ./

# Build the React application
RUN npm run build

# Backend build stage
FROM golang:1.24-alpine AS backend-builder

# Install minimal dependencies for CGO-free builds
RUN apk add --no-cache git ca-certificates tzdata

# Set CGO environment variables for static builds
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/server ./cmd/server

# Build migration utility with static linking
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/migrate ./debug/fix_production_schema.go

# Build the Railway migration runner with static linking
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/railway_migration_runner ./debug/railway_migration_runner.go

# Final stage
FROM alpine:latest

# Install runtime dependencies including bash
RUN apk add --no-cache ca-certificates tzdata wget bash

# Create app directory
RUN mkdir -p /app

# Copy binary from backend builder
COPY --from=backend-builder /app/bin/server /app/server

# Copy migration utility, Railway migration runner, and SQL script
COPY --from=backend-builder /app/bin/migrate /app/migrate
COPY --from=backend-builder /app/bin/railway_migration_runner /app/railway_migration_runner
COPY --from=backend-builder /src/production_fix_jam_column.sql /app/production_fix_jam_column.sql

# Copy startup script
COPY --from=backend-builder /src/start-with-migration.sh /app/start-with-migration.sh
RUN chmod +x /app/start-with-migration.sh

# Copy built React application from frontend builder
COPY --from=frontend-builder /app/dist /app/dist

# Copy templates and static files (fallback)
COPY --from=backend-builder /src/templates /app/templates
COPY --from=backend-builder /src/static /app/static

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Set default environment variables (Railway will override PORT at runtime)
ENV PORT=8080
ENV APP_ENV=production

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/healthz || exit 1

# Run the application with migration
CMD ["/app/start-with-migration.sh"]