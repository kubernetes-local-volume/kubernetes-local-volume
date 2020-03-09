#!/usr/bin/env bash

set -o errexit
set -o nounset

# work dir
export WORK_DIR=$(cd `dirname $0`; cd ..; pwd)
mkdir -p ${WORK_DIR} || true

# output dir
export OUTPUT_DIR=${WORK_DIR}/build/_output
mkdir -p ${OUTPUT_DIR} || true

rm -rf ${OUTPUT_DIR}
