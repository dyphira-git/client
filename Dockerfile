# Build stage
FROM golang:1.24.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/dyp_chain

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/dyp_chain /app/dyp_chain

# Create data directory for blockchain and wallet
RUN mkdir -p /app/data

# Set environment variables
ENV GIN_MODE=release

# Expose ports (adjust as needed based on your application)
EXPOSE 8080
EXPOSE 50051

# Command to run the application
CMD ["/app/dyp_chain"] 