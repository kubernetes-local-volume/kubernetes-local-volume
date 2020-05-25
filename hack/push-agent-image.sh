#!/usr/bin/env bash

set -o errexit
set -o nounset

IMAGE=localvolume/local-volume-csi-agent:latest

# work dir
export WORK_DIR=$(cd `dirname $0`; cd ..; pwd)
mkdir -p ${WORK_DIR} || true

# push image
docker push ${IMAGE}
