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

package driver

import (
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
)

type identityServer struct {
	driver *LocalVolumeDriver
	*csicommon.DefaultIdentityServer
}

// newIdentityServer create identity server
func newIdentityServer(d *LocalVolumeDriver) *identityServer {
	return &identityServer{
		driver:                d,
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d.csiDriver),
	}
}

func (iden *identityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	logging.GetLogger().Infof("Identity:GetPluginInfo Request :: %+v", *req)

	if iden.driver.driverName == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if iden.driver.driverVersion == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          iden.driver.driverName,
		VendorVersion: iden.driver.driverVersion,
	}, nil
}

func (iden *identityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	logging.GetLogger().Infof("Identity:Probe Request :: %+v", *req)
	return &csi.ProbeResponse{}, nil
}

// GetPluginCapabilities returns available capabilities of the plugin
func (iden *identityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	logging.GetLogger().Infof("Identity:GetPluginCapabilities Request :: %+v", *req)
	resp := &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
			{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_OFFLINE,
					},
				},
			},
		},
	}
	return resp, nil
}
