#!/usr/bin/env bash

set -o errexit
set -o nounset

# work dir
export WORK_DIR=$(cd `dirname $0`; cd ..; pwd)
mkdir -p ${WORK_DIR} || true

kubectl apply -f ${WORK_DIR}/examples/storageclass.yaml
kubectl apply -f ${WORK_DIR}/examples/pvc.yaml
kubectl apply -f ${WORK_DIR}/examples/deploy.yaml
