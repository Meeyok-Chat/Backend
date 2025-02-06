FROM golang:1.23.2

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy  # ทำการติดตั้ง dependencies ที่จำเป็น

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping cmd/main.go

EXPOSE 8000

CMD ["/docker-gs-ping"]