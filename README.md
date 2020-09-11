# [WIP] Prometheus SQL Remote Storage Adapter for Generic RDBMS

 - [![Build on Linux](https://github.com/ledyba/prometheus_sql_adapter/workflows/Build%20on%20Linux/badge.svg)](https://github.com/ledyba/prometheus_sql_adapter/actions?query=workflow%3A%22Build+on+Linux%22)
   - [![Build single binary on Linux](https://github.com/ledyba/prometheus_sql_adapter/workflows/Build%20single%20binary%20on%20Linux/badge.svg)](https://github.com/ledyba/prometheus_sql_adapter/actions?query=workflow%3A%22Build+single+binary+on+Linux%22)
 - [![Build on macOS](https://github.com/ledyba/prometheus_sql_adapter/workflows/Build%20on%20macOS/badge.svg)](https://github.com/ledyba/prometheus_sql_adapter/actions?query=workflow%3A%22Build+on+macOS%22)
 - [![Build on Windows](https://github.com/ledyba/prometheus_sql_adapter/workflows/Build%20on%20Windows/badge.svg)](https://github.com/ledyba/prometheus_sql_adapter/actions?query=workflow%3A%22Build+on+Windows%22)

Prometheus remote storage adapter, which stores timeseries data into RDBMS.

## Building and running

### with Cargo

```bash
cargo build --release
```

then run,

```bash
target/release/prometheus_sql_adapter web \
  --listen '0.0.0.0:8080' \
  --db '....' # TODO
```

### with Docker

Write a docker-compose.yml like:

```yaml
---
version: '3.7'

services:
  prometheus_sql_adapter:
    container_name: prometheus_sql_adapter
    hostname: prometheus_sql_adapter
    image: prometheus_sql_adapter
    build:
      context: ./
    restart: always
    command:
      - 'web'
      - '--listen'
      - '0.0.0.0:8080'
      - '--db'
 # to use sqlite,
      - 'sqlite:///var/lib/sqlite/prom.db'
#     - 'sqlite:' # use in-memory db(for debugging)
      - '--init' # init db before launching
```

then,

```bash
docker-comopse build # It takes long time. Be patient....
docker-comopse up -d
```

## Using from Prometheus

Write these line to `/etc/prometheus/config.yml`

```yaml
remote_write:
  - url: 'http://<hostname>:8080/write'
remote_read:
  - url: 'http://<hostname>:8080/read'
```
