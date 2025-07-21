# Production Dockerfile
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./cmd/gateway

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S mmorpg && \
    adduser -u 1000 -S mmorpg -G mmorpg

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gateway .

# Copy config files if needed
# COPY --from=builder /build/config ./config

# Change ownership
RUN chown -R mmorpg:mmorpg /app

# Switch to non-root user
USER mmorpg

# Expose port
EXPOSE 8090

# Run the application
CMD ["./gateway"]