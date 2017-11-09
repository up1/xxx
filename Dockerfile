FROM golang:1.9.0-alpine3.6 as builder
WORKDIR /go/src/lifestyle
COPY . .
RUN go test -v --cover ./...
