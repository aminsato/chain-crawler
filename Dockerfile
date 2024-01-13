
FROM golang:1.21-alpine AS build

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY . .



RUN CGO_ENABLED=0  go build -o /crw ./cmd/main.go

#docker run  --name eth-crw  -p 1080:1080 crw  /ethereum-crawler  eth  --http-port 1080 docker ps -a

# Run
CMD ["/crw"]

