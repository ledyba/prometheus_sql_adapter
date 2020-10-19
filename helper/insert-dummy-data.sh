#! /bin/bash -eu

PROJ_PATH=$(cd "$(dirname "$(readlink -f "$0")")" && pwd)
cd "${PROJ_PATH}" || exit

set -x

./avalanche \
  --value-interval=10 --metric-count=10 --label-count=3 --series-count=3 --remote-batch-size=300 \
  --remote-requests-count=100 \
  --remote-url=http://localhost:8080/write
