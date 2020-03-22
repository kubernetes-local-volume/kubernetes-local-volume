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

.PHONY: build
build: build-driver build-scheduler

.PHONY: build-driver
build-driver:
	./hack/build.sh driver local.volume.csi.driver.kubernetes.io

.PHONY: build-scheduler
build-scheduler:
	./hack/build.sh scheduler local.volume.csi.scheduler.kubernetes.io

.PHONY: make-driver-image
make-driver-image: build
	./hack/make-driver-image.sh

.PHONY: push-driver-image
push-driver-image: make-driver-image
	./hack/push-driver-image.sh

.PHONY: deploy-driver
deploy-driver:
	./hack/deploy-driver.sh

.PHONY: undeploy-driver
undeploy-driver:
	./hack/undeploy-driver.sh

.PHONY: start-driver-test
start-driver-test:
	./hack/start-driver-test.sh

.PHONY: stop-driver-test
stop-driver-test:
	./hack/stop-driver-test.sh

.PHONY: generate
generate:
	hack/codegen/codegen.sh

.PHONY: clean
clean:
	./hack/clean.sh
