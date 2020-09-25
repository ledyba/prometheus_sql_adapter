.PHONY: all clean test

all: prometheus_sql_adapter;

SRCS := $(shell find . -type f -name *.go)

prometheus_sql_adapter: $(SRCS)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o prometheus_sql_adapter ./cmd/prometheus_sql_adapter

clean:
	rm -f prometheus_sql_adapter

test:
	GOOS=linux GOARCH=amd64 go test ./...
