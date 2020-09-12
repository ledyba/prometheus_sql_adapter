#! /bin/bash -eu

PROJ_PATH=$(readlink -f $(cd $(dirname $(readlink -f $0)) && pwd))
cd ${PROJ_PATH}/..

mkdir -p artifacts

docker-compose build
docker cp prometheus_sql_adapter:/prometheus_sql_adapter ${PWD}/artifacts/prometheus_sql_adapter
