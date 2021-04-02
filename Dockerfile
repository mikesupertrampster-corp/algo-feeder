FROM golang:1.16.2 as builder

WORKDIR /workspace
COPY . .
RUN env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

FROM alpine:3.13.3
COPY --from=builder /workspace/ /

ENTRYPOINT ["/main"]