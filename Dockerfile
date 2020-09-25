FROM golang:1.14-alpine as build

WORKDIR /go/src/github.com/ledyba/prometheus_sql_adapter
COPY --chown=rust:rust . .

RUN apk add git gcc g++ musl-dev bash make sqlite-dev mysql-client

RUN make clean && make

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
