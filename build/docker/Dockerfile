# syntax=docker/dockerfile:1
FROM alpine:golang:1.18.5 as builer
ENV GO111MODULE=on
WORKDIR /kurajj
COPY ../../../.. .
RUN apk --no-cache add git alpine-sdk build-base gcc
RUN go mod download
RUN go build -o kurajj cmd/kurajj/main.go

FROM alpine:3.15.4

RUN apk --no-cache add ca-certificates

WORKDIR /root

COPY --from=builder /kurajj/kurajj .

CMD ["./storage"]
