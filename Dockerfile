# Stage 1: Build Go binary
FROM golang:1.23.2 AS builder

WORKDIR /app

# Install swag CLI tool
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go.mod and go.sum to download dependencies
# COPY go.mod ./

# Download dependencies


# Copy the entire project
COPY . ./

RUN go mod tidy

# Generate Swagger documentation inside the cmd/docs directory
RUN swag init -g cmd/main.go -o cmd/docs

# Build the Go application binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/docker-gs-ping ./cmd/main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/docker-gs-ping .

# Copy Swagger documentation to runtime container (optional)
COPY --from=builder /app/cmd/docs /root/docs

# Grant execute permission to the binary
RUN chmod +x /root/docker-gs-ping

EXPOSE 8000

CMD ["/root/docker-gs-ping"]