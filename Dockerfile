# Build stage
FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY ./engine/go.mod ./engine/go.sum ./

# Download dependencies
RUN go mod download

# Copy the application source code
COPY ./engine .

# Build the Go application
RUN go build -o api-server .

# Runtime stage
FROM alpine:3.20

# Set the working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/api-server .

# Set Proper permissions for TLS private key (not needed for local development)
# RUN chmod 600 ./tls/privkey.pem 

# Expose the application port
EXPOSE 1234

# Run the application
CMD ["./api-server"]
