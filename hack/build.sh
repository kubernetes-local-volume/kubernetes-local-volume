#!/usr/bin/env bash

set -o errexit
set -o nounset

# work dir
export WORK_DIR=$(cd `dirname $0`; cd ..; pwd)
mkdir -p ${WORK_DIR} || true

# output dir
export OUTPUT_DIR=${WORK_DIR}/build/_output
mkdir -p ${OUTPUT_DIR} || true

# build function
go_build () {
	echo "[START] building "kubernetes local volume component $1"..."
	# We’re disabling cgo which gives us a static binary.
	# This is needed for building minimal container based on alpine image.
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${OUTPUT_DIR}/$1 -installsuffix cgo -ldflags "$go_ldflags" ${WORK_DIR}/cmd/$1/
	echo "[END] building "kubernetes local volume component $1"..."
}

# check golang
if ! which go > /dev/null; then
	echo "golang needs to be installed"
	exit 1
fi

GIT_SHA=`git rev-parse --short HEAD || echo "GitNotFound"`

gitHash="github.com/kubernetes-local-volume/kubernetes-local-volume/version.GitSHA=${GIT_SHA}"

go_ldflags="-X ${gitHash}"

go_build $@
