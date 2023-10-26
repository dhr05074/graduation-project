FROM golang:1.18 AS builder

WORKDIR /source

ENV GOFLAGS="-buildvcs=false"

RUN go mod download
RUN go build -o traffic.so -buildmode=c-shared .
