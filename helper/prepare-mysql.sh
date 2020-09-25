#! /bin/bash -eu

PROJ_PATH=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)
cd "${PROJ_PATH}/.." || exit

docker-compose up -d mysql
sleep 5
echo "create database db"
docker-compose down
