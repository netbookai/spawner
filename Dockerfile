FROM golang:1.17-alpine as builder
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /spawnerservice

# Copying code for build
COPY go.mod ./
COPY go.sum ./
COPY config.env ./
COPY cmd ./cmd
COPY pb ./pb
COPY pkg ./pkg

WORKDIR /spawnerservice/cmd/spawnersvc

# Optimized build by removing debug info and compile only for linux target and disabling compilation.
RUN go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/spawnersvc 

WORKDIR /spawnerservice/cmd/spawnercli

#Optimized build by removing debug info and compile only for linux target and disabling compilation.
RUN go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/spawnercli

FROM alpine  
COPY config.env ./
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt 
COPY --from=builder /go/bin/spawnersvc /go/bin/spawnersvc
COPY --from=builder /go/bin/spawnercli /go/bin/spawnercli
