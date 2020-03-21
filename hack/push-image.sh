#!/usr/bin/env bash

set -o errexit
set -o nounset

IMAGE=core.harbor.domain/webhook/local-volume-csi-plugin:latest

# work dir
export WORK_DIR=$(cd `dirname $0`; cd ..; pwd)
mkdir -p ${WORK_DIR} || true

# push image
docker push ${IMAGE}
