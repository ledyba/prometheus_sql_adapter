# [WIP] Prometheus SQL Remote Storage Adapter for Generic RDBMS

[![Build on Linux](https://github.com/ledyba/prometheus_sql_adapter/workflows/Build%20on%20Linux/badge.svg)](https://github.com/ledyba/prometheus_sql_adapter/actions?query=workflow%3A%22Build+on+Linux%22)
[![Build on macOS](https://github.com/ledyba/prometheus_sql_adapter/workflows/Build%20on%20macOS/badge.svg)](https://github.com/ledyba/prometheus_sql_adapter/actions?query=workflow%3A%22Build+on+macOS%22)
[![Build on Windows](https://github.com/ledyba/prometheus_sql_adapter/workflows/Build%20on%20Windows/badge.svg)](https://github.com/ledyba/prometheus_sql_adapter/actions?query=workflow%3A%22Build+on+Windows%22)  
[![Build single binary on Linux](https://github.com/ledyba/prometheus_sql_adapter/workflows/Build%20single%20binary%20on%20Linux/badge.svg)](https://github.com/ledyba/prometheus_sql_adapter/actions?query=workflow%3A%22Build+single+binary+on+Linux%22)
 

Prometheus remote storage adapter, which stores timeseries data into RDBMS.

## Building and running

### with Makefile

You need to install golang for building.

```bash
make
```

then run,

```bash
./prometheus_sql_adapter web \
    --listen '0.0.0.0:8080' \
 # to use sqlite,
    --db 'sqlite://file:/var/lib/sqlite/prometheus.db?cache=shared&mode=rwc'
#   --db 'sqlite://file::memory:?cache=shared' # use in-memory db(for debugging)
# to use mysql,
#   --db 'mysql://root:root@tcp(mysql-server-addr:3306)/db'

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
      - 'sqlite://file:/var/lib/sqlite/prometheus.db?cache=shared&mode=rwc'
#     - 'sqlite://file::memory:?cache=shared' # use in-memory db(for debugging)
# to use mysql,
#     - 'mysql://root:root@tcp(mysql-server-addr:3306)/db'
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
