############################
# STEP 1 build executable binary
# Updated: 2025-06-24 - Force rebuild with auth fixes
# Cache bust: v1.1.0-fixed
############################
FROM golang:1.23-alpine3.20 AS builder
RUN apk update && apk add --no-cache gcc musl-dev gcompat
WORKDIR /whatsapp

# Copy the source code
COPY ./src .

# Debug: List files to verify dashboard.html is copied
RUN ls -la views/dashboard.html || echo "dashboard.html not found!"

# Fetch dependencies.
RUN go mod download
# Build the binary with optimizations
RUN go build -a -ldflags="-w -s" -o /app/whatsapp

#############################
## STEP 2 build a smaller image
#############################
FROM alpine:3.20
RUN apk add --no-cache ffmpeg ca-certificates
WORKDIR /app

# Copy compiled from builder.
COPY --from=builder /app/whatsapp /app/whatsapp

# Create necessary directories
RUN mkdir -p /app/storages /app/statics/qrcode /app/statics/media /app/statics/senditems

# Expose port (Railway will override this with PORT env var)
EXPOSE 3000

# Run the binary.
ENTRYPOINT ["/app/whatsapp"]
CMD [ "rest" ]