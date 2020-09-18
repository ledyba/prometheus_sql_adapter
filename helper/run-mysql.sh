#! /bin/bash -eu

PROJ_PATH=$(cd $(dirname $(readlink -f $0)) && pwd)
cd ${PROJ_PATH}/..

docker-compose up -d mysql
docker-compose exec mysql mysql \
  --host localhost --port 3306 --protocol=TCP \
  --user=root --password=root
docker-compose down
