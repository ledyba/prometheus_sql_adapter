#! /bin/bash -eu

PROJ_PATH=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)
cd "${PROJ_PATH}/.." || exit

set -x

docker-compose up -d mysql
while ! docker-compose exec mysql bash -c "mysql --user=root --password=root --host=localhost --port=3306 --execute 'drop database db; create database db'"; do
  echo "Retry..."
done
docker-compose down
