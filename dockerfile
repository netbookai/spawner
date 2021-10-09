FROM golang:1.17-alpine

WORKDIR /spawnerservice

COPY go.mod ./
COPY go.sum ./
COPY config.env ./
COPY cmd ./cmd
COPY pb ./pb
COPY pkg ./pkg

WORKDIR /spawnerservice/cmd/spawnersvc

RUN go build