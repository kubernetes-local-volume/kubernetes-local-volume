#!/usr/bin/env bash

# generate-groups.sh  -> https://github.com/kubernetes/code-generator/blob/master/generate-groups.sh

# work dir
export WORK_DIR=$(cd `dirname $0`; cd ../..; pwd)

# install kubernetes code generator
# go get -u k8s.io/code-generator/...

# generate deepcopy, client, informer, lister by kubernetes code generator
bash ${WORK_DIR}/hack/codegen/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/apis \
  "storage:v1alpha1" \
  --go-header-file ${WORK_DIR}/hack/codegen/boilerplate.go.txt