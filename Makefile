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
	./hack/build.sh driver local.volume.csi.kubernetes.io

.PHONY: build-scheduler
build-scheduler:
	./hack/build.sh scheduler local.volume.scheduler.kubernetes.io

.PHONY: image
image: build
	./hack/make-image.sh

.PHONY: push
push: image
	./hack/push-image.sh

.PHONY: deploy
deploy:
	./hack/deploy.sh

.PHONY: undeploy
undeploy:
	./hack/undeploy.sh

.PHONY: start-test
start-test:
	./hack/start-test.sh

.PHONY: stop-test
stop-test:
	./hack/stop-test.sh

.PHONY: generate
generate:
	hack/codegen/codegen.sh

.PHONY: clean
clean:
	./hack/clean.sh
