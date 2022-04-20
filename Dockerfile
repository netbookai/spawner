FROM golang:1.17.5-alpine as builder
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
COPY proto ./proto
COPY pkg ./pkg

WORKDIR /spawnerservice/cmd/spawnersvc

# Optimized build by removing debug info and compile only for linux target and disabling compilation.
RUN go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/spawnersvc 

WORKDIR /spawnerservice/cmd/spawnercli


FROM alpine  
COPY config.env ./
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt 
COPY --from=builder /go/bin/spawnersvc /go/bin/spawnersvc
