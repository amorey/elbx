FROM golang:1.16.3 AS builder

RUN mkdir app
WORKDIR /app

# install dependencies first (for cache)
COPY go.mod .
COPY go.sum .
RUN go mod download

# copy code
COPY . .

# build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main

ENTRYPOINT ["./main"]

# -----------------------------------------------------------------------------

FROM alpine:latest as alpine

RUN apk add -U --no-cache ca-certificates

# -----------------------------------------------------------------------------

FROM scratch

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app /app
WORKDIR /app

ENTRYPOINT ["./main"]
