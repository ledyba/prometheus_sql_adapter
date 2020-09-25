SRCS := $(shell find . -type f -name *.go)

.PHONY: all
all: prometheus_sql_adapter;

prometheus_sql_adapter: $(SRCS)
	CGO_ENABLED=1 go build -o prometheus_sql_adapter ./cmd/prometheus_sql_adapter

.PHONY: clean
clean:
	rm -f prometheus_sql_adapter

.PHONY: musl-static
musl-static:
	CGO_ENABLED=1 CC=/usr/bin/musl-gcc go build --ldflags '-linkmode external -extldflags "-static"' \
		-o prometheus_sql_adapter ./cmd/prometheus_sql_adapter

.PHONY: test
test:
	go test ./...
