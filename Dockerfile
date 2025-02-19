# Stage 1: Build Go binary
FROM golang:1.23.2 AS builder

WORKDIR /app

# Copy และติดตั้ง dependencies
COPY go.mod go.sum ./
RUN go mod tidy 

COPY . ./

# Build binary ไปไว้ที่ /app/docker-gs-ping
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/docker-gs-ping ./cmd/main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /root/

# Copy binary จาก builder stage
COPY --from=builder /app/docker-gs-ping .

# ให้สิทธิ์ execute กับไฟล์
RUN chmod +x /root/docker-gs-ping

EXPOSE 8000

CMD ["/root/docker-gs-ping"]