#! /bin/bash -eu

PROJ_PATH=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)
cd "${PROJ_PATH}/.." || exit

docker-compose up -d mysql
sleep 5
docker-compose exec mysql bash -c "mysql"
docker-compose down
