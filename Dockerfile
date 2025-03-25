# Use the official Golang image
FROM golang:1.23.2

# Install swag CLI tool
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the entire project
COPY . ./

# Generate Swagger documentation in cmd/ directory
RUN swag init -g cmd/main.go -o cmd/docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping cmd/main.go

# Expose the port the app runs on
EXPOSE 8000

# Command to run the executable
CMD ["/docker-gs-ping"]