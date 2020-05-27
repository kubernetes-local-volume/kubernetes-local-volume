#!/usr/bin/env bash

set -o errexit
set -o nounset

# work dir
export WORK_DIR=$(cd `dirname $0`; cd ..; pwd)
mkdir -p ${WORK_DIR} || true

kubectl delete -f ${WORK_DIR}/deploy/local-volume-csi.yaml
kubectl delete -f ${WORK_DIR}/deploy/local-volume-provisioner.yaml
kubectl delete -f ${WORK_DIR}/deploy/local-volume-rbac.yaml
kubectl delete -f ${WORK_DIR}/deploy/local-volume-crd.yaml
