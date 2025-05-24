# Start from the official Golang image
FROM golang:latest

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

# Expose the application port
EXPOSE 443

# Run the application
CMD ["./api-server"]