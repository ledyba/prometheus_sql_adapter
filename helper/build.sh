#! /bin/bash -eu

PROJ_PATH=$(cd "$(dirname "$(readlink -f $0)")" && pwd)
cd "${PROJ_PATH}/.."

mkdir -p artifacts

docker network create planet-link
docker-compose build
docker-compose up -d
docker cp prometheus_sql_adapter:/prometheus_sql_adapter "${PWD}/artifacts/prometheus_sql_adapter"
docker-compose down
