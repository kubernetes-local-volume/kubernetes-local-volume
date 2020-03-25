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
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
)

type controllerServer struct {
	driver *LocalVolumeDriver
	*csicommon.DefaultControllerServer
}

// newControllerServer creates a controllerServer object
func newControllerServer(d *LocalVolumeDriver) *controllerServer {
	return &controllerServer{
		driver:                  d,
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.csiDriver),
	}
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	logging.GetLogger().Infof("Controller:CreateVolume Request :: %+v", *req)

	if err := cs.driver.csiDriver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		logging.GetLogger().Infof("invalid create volume req: %v", *req)
		return nil, err
	}
	if len(req.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume Name cannot be empty")
	}
	if req.VolumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities cannot be empty")
	}

	// Get nodeID if pvc in topology mode.
	nodeID := pickNodeID(req.GetAccessibilityRequirements())
	if nodeID == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeID cannot be empty")
	}

	response := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      req.GetName(),
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext: req.GetParameters(),
			AccessibleTopology: []*csi.Topology{
				{
					Segments: map[string]string{
						types.TopologyNodeKey: nodeID,
					},
				},
			},
		},
	}

	logging.GetLogger().Infof("Controller:CreateVolume Success :: volume = %s, size = %d", req.GetName(), req.GetCapacityRange().GetRequiredBytes())
	return response, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	logging.GetLogger().Infof("Controller:DeleteVolume Request :: %+v", *req)
	logging.GetLogger().Infof("Controller:DeleteVolume Success :: volume = %s", req.GetVolumeId())
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	logging.GetLogger().Infof("Controller:ControllerExpandVolume Request :: %+v", *req)
	volSizeBytes := int64(req.GetCapacityRange().GetRequiredBytes())
	logging.GetLogger().Infof("Controller:ControllerExpandVolume Success :: volume = %s", req.GetVolumeId())
	return &csi.ControllerExpandVolumeResponse{CapacityBytes: volSizeBytes, NodeExpansionRequired: true}, nil
}

// pickNodeID selects node given topology requirement.
// if not found, empty string is returned.
func pickNodeID(requirement *csi.TopologyRequirement) string {
	if requirement == nil {
		return ""
	}
	for _, topology := range requirement.GetPreferred() {
		nodeID, exists := topology.GetSegments()[types.TopologyNodeKey]
		if exists {
			return nodeID
		}
	}
	for _, topology := range requirement.GetRequisite() {
		nodeID, exists := topology.GetSegments()[types.TopologyNodeKey]
		if exists {
			return nodeID
		}
	}
	return ""
}
