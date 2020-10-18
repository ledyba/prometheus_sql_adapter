#! /bin/bash -eu

PROJ_PATH=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)
cd "${PROJ_PATH}" || exit

set -x

./avalanche --value-interval=10 --remote-url=http://localhost:8080/write
