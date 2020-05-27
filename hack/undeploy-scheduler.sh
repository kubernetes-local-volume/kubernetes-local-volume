#!/usr/bin/env bash

set -o errexit
set -o nounset

# work dir
export WORK_DIR=$(cd `dirname $0`; cd ..; pwd)
mkdir -p ${WORK_DIR} || true

kubectl delete -f ${WORK_DIR}/deploy/local-volume-scheduler.yaml
