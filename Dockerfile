
FROM golang:1.21-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .



RUN CGO_ENABLED=0  go build -o /crw ./crw/main.go

CMD ["/crw"]

