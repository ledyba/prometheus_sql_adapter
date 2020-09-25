FROM golang:1.15-alpine as builder

WORKDIR /go/src/github.com/ledyba/prometheus_sql_adapter
COPY . .

RUN apk add git gcc g++ musl-dev bash make sqlite-dev mysql-client

ENV GOOS=linux
ENV GOARCH=amd64
RUN make clean && make musl-static

FROM alpine:3.12

WORKDIR /

RUN apk add --no-cache ca-certificates && update-ca-certificates

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV SSL_CERT_DIR=/etc/ssl/certs

COPY --chown=nobody:nogroup --from=builder /go/src/github.com/ledyba/prometheus_sql_adapter/prometheus_sql_adapter prometheus_sql_adapter
RUN ["chmod", "a+x", "/prometheus_sql_adapter"]
EXPOSE 8080
USER nobody
ENTRYPOINT ["/prometheus_sql_adapter"]
