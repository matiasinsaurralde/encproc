# Start from the official Golang image
FROM golang:1.22

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application
RUN go build -o api-server .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./api-server"]