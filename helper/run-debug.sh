#! /bin/bash -eu

PROJ_PATH=$(cd $(dirname $(readlink -f $0)) && pwd)
cd ${PROJ_PATH}/..

trap 'docker-compose down' 2
trap 'docker-compose down' 4

set -x
docker-compose run local || docker-compose down

