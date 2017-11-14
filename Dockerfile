FROM golang:1.9.0-alpine3.6 as builder
WORKDIR /go/src/lifestyle
COPY . .
RUN go test -v --cover ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lifestyle .

#FROM alpine:latest
#RUN apk --no-cache add ca-certificates
#WORKDIR /root/
#COPY --from=builder /go/src/lifestyle
#CMD ["./lifestyle"]
