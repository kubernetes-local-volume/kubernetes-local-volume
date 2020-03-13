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
build: build-lvm build-scheduler

.PHONY: build-lvm
build-lvm:
	./build/build.sh lvm local.volume.csi.kubernetes.io

.PHONY: build-scheduler
build-scheduler:
	./build/build.sh scheduler local.volume.scheduler.kubernetes.io

.PHONY: image
image: build
	./build/make-image.sh

.PHONY: push
push: image
	./build/push-image.sh

.PHONY: clean
clean:
	./hack/clean.sh
