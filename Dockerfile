FROM golang:alpine as builder
WORKDIR /build
ADD . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /go/bin/go-relay main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/bin/go-relay /usr/bin/go-relay

CMD [ "go-relay", "-h" ]