FROM golang:alpine AS builder

ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk update --no-cache

WORKDIR /build

COPY . .

RUN go build -ldflags="-s -w" -o main cmd/main.go

FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates

WORKDIR /build

COPY --from=builder /build/main /build/main

CMD ["./main"]