############ Build stage ############
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Go caching tip: copy the go.mod first so `go mod download` can cache layers
COPY engine/go.mod engine/go.sum ./
RUN go mod download

# Copy the rest of the source
COPY engine .

# Build a small statically linked binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/api-server .

############ Runtime stage ############
FROM alpine:3.20

# Create an unprivileged user (1000:1000 matches compose)
RUN addgroup -g 1000 app && adduser -S -u 1000 -G app app

WORKDIR /app
COPY --from=builder /out/api-server .         

EXPOSE 8080  
EXPOSE 9000  
EXPOSE 443    

USER 1000:1000
CMD ["./api-server"]
