# syntax=docker/dockerfile:1
FROM golang:1.18.5-alpine3.15 as builder
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
WORKDIR /kurajj
COPY ../../../.. .
RUN go mod download
RUN go build -ldflags="-s -w" -o kurajj cmd/kurajj/main.go

FROM alpine:3.15.4

RUN apk --no-cache add ca-certificates

WORKDIR /root

COPY --from=builder /kurajj/kurajj .
COPY --from=builder /kurajj/configs/ configs/
COPY --from=builder /kurajj/tls/ tls/
COPY --from=builder /kurajj/internal/templates/ internal/templates/

CMD ["./kurajj"]
