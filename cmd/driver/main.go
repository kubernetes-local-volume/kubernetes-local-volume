/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/driver"
)

var (
	endpoint = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID   = flag.String("nodeid", "", "node id")
)

// Nas CSI Plugin
func main() {
	flag.Parse()

	driver := driver.NewLocalVolumeDriver(types.DriverName, types.DriverVersion, *nodeID, *endpoint)
	driver.Run()

	os.Exit(0)
}
