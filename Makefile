# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# build all component
.PHONY: build
build: build-driver build-scheduler build-agent

.PHONY: build-driver
build-driver:
	./hack/build.sh driver local.volume.csi.driver.kubernetes.io

.PHONY: build-scheduler
build-scheduler:
	./hack/build.sh scheduler local.volume.csi.scheduler.kubernetes.io

.PHONY: build-agent
build-agent:
	./hack/build.sh agent local.volume.csi.agent.kubernetes.io

# image
.PHONY: make-image
make-image: make-driver-image make-agent-image

.PHONY: push-image
push-image: push-driver-image push-agent-image

.PHONY: make-driver-image
make-driver-image: build-driver
	./hack/make-driver-image.sh

.PHONY: push-driver-image
push-driver-image: make-driver-image
	./hack/push-driver-image.sh

.PHONY: make-agent-image
make-agent-image: build-agent
	./hack/make-agent-image.sh

.PHONY: push-agent-image
push-agent-image: make-agent-image
	./hack/push-agent-image.sh

# deploy
.PHONY: deploy
deploy:
	./hack/deploy.sh

.PHONY: undeploy
undeploy:
	./hack/undeploy.sh

# test
.PHONY: start-test
start-test:
	./hack/start-test.sh

.PHONY: stop-test
stop-test:
	./hack/stop-test.sh

# generate crd sdk
.PHONY: generate
generate:
	hack/codegen/codegen.sh

.PHONY: clean
clean:
	./hack/clean.sh
