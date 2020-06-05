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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/util/resizefs"
	k8sexec "k8s.io/utils/exec"
	k8smount "k8s.io/utils/mount"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/lvm"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/mounter"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/utils"
)

const (
	// FsTypeTag is the fs type tag
	FsTypeTag = "fsType"
	// LvmTypeTag is the lvm type tag
	LvmTypeTag = "lvmType"
	// LinearType linear type
	LinearType = "linear"
	// StripingType striping type
	StripingType = "striping"
	// DefaultFs default fs
	DefaultFs = "ext4"
)

const (
	volumePublishSuccess = "local.volume.csi.kubernetes.io/publish"
)

type nodeServer struct {
	driver *LocalVolumeDriver
	*csicommon.DefaultNodeServer
	nodeID     string
	mounter    mounter.Mounter
	client     kubernetes.Interface
	k8smounter k8smount.Interface
}

var (
	masterURL  string
	kubeconfig string
)

// NewNodeServer create a NodeServer object
func NewNodeServer(d *LocalVolumeDriver, nodeID string) csi.NodeServer {
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logging.GetLogger().Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logging.GetLogger().Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	return &nodeServer{
		driver:            d,
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d.csiDriver),
		nodeID:            nodeID,
		mounter:           mounter.NewMounter(),
		k8smounter:        k8smount.New(""),
		client:            kubeClient,
	}
}

func (ns *nodeServer) GetNodeID() string {
	return ns.nodeID
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	logging.GetLogger().Infof("NodeServer:NodePublishVolume Request :: %+v", *req)

	// parse request args.
	targetPath := req.GetTargetPath()
	if targetPath == "" {
		return nil, status.Error(codes.Internal, "targetPath is empty")
	}
	lvmType := LinearType
	if _, ok := req.VolumeContext[LvmTypeTag]; ok {
		lvmType = req.VolumeContext[LvmTypeTag]
	}
	fsType := DefaultFs
	if _, ok := req.VolumeContext[FsTypeTag]; ok {
		fsType = req.VolumeContext[FsTypeTag]
	}
	logging.GetLogger().Infof("NodeServerNodePublishVolume :: Starting to mount lvm at %s, with vg %s, with volume = %s, LVM type = %s",
		targetPath, types.VGName, req.GetVolumeId(), lvmType)

	volumeNewCreated := false
	volumeID := req.GetVolumeId()
	devicePath := filepath.Join("/dev/", types.VGName, volumeID)
	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		volumeNewCreated = true
		err := ns.createVolume(ctx, volumeID, types.VGName, lvmType)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	isMnt, err := ns.mounter.IsMounted(targetPath)
	if err != nil {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			isMnt = false
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	exitFSType, err := checkFSType(devicePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check fs type err: %v", err)
	}
	if exitFSType == "" {
		logging.GetLogger().Infof("The device %v has no filesystem, starting format: %v", devicePath, fsType)
		if err := formatDevice(devicePath, fsType); err != nil {
			return nil, status.Errorf(codes.Internal, "format fstype failed: err=%v", err)
		}
	}

	if !isMnt {
		var options []string
		if req.GetReadonly() {
			options = append(options, "ro")
		} else {
			options = append(options, "rw")
		}
		mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
		options = append(options, mountFlags...)

		err = ns.mounter.Mount(devicePath, targetPath, fsType, options...)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		logging.GetLogger().Infof("NodeServer:NodePublishVolume Success :: mount successful devicePath = %s, targetPath = %s, options = %v",
			devicePath, targetPath, options)
	}

	// xfs filesystem works on targetpath.
	if volumeNewCreated == false {
		if err := ns.resizeVolume(ctx, volumeID, types.VGName, targetPath); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Update PersistentVolume tag, inform agent controller update localvolume free size
	if err := ns.updatePVPublishSuccessTag(ctx, volumeID); err != nil {
		logging.GetLogger().Errorf("NodeServer:NodePublishVolume update PV publish success tag error : %+v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) updatePVPublishSuccessTag(ctx context.Context, volumeID string) error {
	oldPv, err := ns.client.CoreV1().PersistentVolumes().Get(volumeID, metav1.GetOptions{})
	if err != nil {
		logging.GetLogger().Errorf("NodePublishVolume: Get Persistent Volume(%s) Error: %s", volumeID, err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	pvClone := oldPv.DeepCopy()
	if pvClone.Annotations == nil {
		pvClone.Annotations = make(map[string]string)
	}

	if _, ok := oldPv.Annotations[volumePublishSuccess]; !ok {
		oldData, err := json.Marshal(oldPv)
		if err != nil {
			logging.GetLogger().Errorf("NodePublishVolume: Marshal Persistent Volume(%s) Error: %s", volumeID, err.Error())
			return status.Error(codes.Internal, err.Error())
		}

		// construct new persistent volume data
		pvClone.Annotations[volumePublishSuccess] = "true"
		newData, err := json.Marshal(pvClone)
		if err != nil {
			logging.GetLogger().Errorf("NodePublishVolume: Marshal New Persistent Volume(%s) Error: %s", volumeID, err.Error())
			return status.Error(codes.Internal, err.Error())
		}
		patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, pvClone)
		if err != nil {
			logging.GetLogger().Errorf("NodePublishVolume: CreateTwoWayMergePatch Volume(%s) Error: %s", volumeID, err.Error())
			return status.Error(codes.Internal, err.Error())
		}

		// Update PersistentVolume
		_, err = ns.client.CoreV1().PersistentVolumes().Patch(volumeID, k8stypes.StrategicMergePatchType, patchBytes)
		if err != nil {
			logging.GetLogger().Errorf("NodePublishVolume: Patch Volume(%s) Error: %s", volumeID, err.Error())
			return status.Error(codes.Internal, err.Error())
		}
		logging.GetLogger().Infof("Update PV(%s) publish success tag success node(%s)", volumeID, ns.nodeID)
	}
	return nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	logging.GetLogger().Infof("NodeServer:NodeUnpublishVolume Request :: %+v", *req)

	targetPath := req.GetTargetPath()
	isMnt, err := ns.mounter.IsMounted(targetPath)
	if err != nil {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return nil, status.Error(codes.NotFound, "TargetPath not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !isMnt {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	err = ns.mounter.Unmount(req.GetTargetPath())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	logging.GetLogger().Infof("NodeServer:NodeUnpublishVolume umount success :: volume = %s, targetPath = %s",
		req.GetVolumeId(), req.GetTargetPath())

	volumeID := req.GetVolumeId()
	devicePath := filepath.Join("/dev/", types.VGName, "/", volumeID)
	logging.GetLogger().Infof("Delete LVM volume, device path: %s", devicePath)
	if _, err := os.Stat(devicePath); err == nil {
		err := ns.deleteVolume(ctx, devicePath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		logging.GetLogger().Errorf("Delete LVM volume, device path: %s, error = %+v", devicePath, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	logging.GetLogger().Infof("NodeServer:NodeUnpublishVolume delete lv(%s) success :: volume = %s",
		devicePath, req.GetVolumeId())

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	logging.GetLogger().Infof("NodeServer:NodeUnstageVolume Request :: %+v", *req)
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	logging.GetLogger().Infof("NodeServer:NodeStageVolume Request :: %+v", *req)
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	logging.GetLogger().Infof("NodeServer:NodeGetCapabilities Request :: %+v", *req)
	// currently there is a single NodeServer capability according to the spec
	nscap := &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			},
		},
	}
	nscap2 := &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
			},
		},
	}
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			nscap, nscap2,
		},
	}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error) {
	logging.GetLogger().Infof("NodeServer:NodeExpandVolume Request :: %+v", *req)
	return &csi.NodeExpandVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	logging.GetLogger().Infof("NodeServer:NodeGetInfo Request :: %+v", *req)
	return &csi.NodeGetInfoResponse{
		NodeId: ns.nodeID,
		// make sure that the driver works on this particular node only
		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{
				types.TopologyNodeKey: ns.nodeID,
			},
		},
	}, nil
}

// create lvm volume
func (ns *nodeServer) createVolume(ctx context.Context, volumeID, vgName, lvmType string) error {
	pvSize, unit := ns.getPvSize(volumeID)

	pvNumber := 0
	var err error
	// Create VG if vg not exist,
	if pvNumber, err = lvm.CreateVG(vgName); err != nil {
		return err
	}

	// check vg exist
	ckCmd := fmt.Sprintf("%s vgck %s", types.NsenterCmd, vgName)
	_, err = utils.Run(ckCmd)
	if err != nil {
		logging.GetLogger().Errorf("createVolume:: VG is not exist: %s", vgName)
		return err
	}

	// Create lvm volume
	if lvmType == StripingType {
		cmd := fmt.Sprintf("%s lvcreate -i %d -n %s -L %d%s %s", types.NsenterCmd, pvNumber, volumeID, pvSize, unit, vgName)
		_, err = utils.Run(cmd)
		if err != nil {
			return err
		}
		logging.GetLogger().Infof("Successful Create Striping LVM volume: %s, Size: %d%s, vgName: %s, striped number: %d", volumeID, pvSize, unit, vgName, pvNumber)
	} else if lvmType == LinearType {
		cmd := fmt.Sprintf("%s lvcreate -n %s -L %d%s %s", types.NsenterCmd, volumeID, pvSize, unit, vgName)
		_, err = utils.Run(cmd)
		if err != nil {
			return err
		}
		logging.GetLogger().Infof("Successful Create Linear LVM volume: %s, Size: %d%s, vgName: %s", volumeID, pvSize, unit, vgName)
	}
	return nil
}

func (ns *nodeServer) resizeVolume(ctx context.Context, volumeID, vgName, targetPath string) error {
	pvSize, unit := ns.getPvSize(volumeID)
	devicePath := filepath.Join("/dev", vgName, volumeID)
	sizeCmd := fmt.Sprintf("%s lvdisplay %s | grep 'LV Size' | awk '{print $3}'", types.NsenterCmd, devicePath)
	sizeStr, err := utils.Run(sizeCmd)
	if err != nil {
		return err
	}
	if sizeStr == "" {
		return status.Error(codes.Internal, "Get lvm size error")
	}
	sizeStr = strings.Split(sizeStr, ".")[0]
	sizeInt, err := strconv.ParseInt(strings.TrimSpace(sizeStr), 10, 64)
	if err != nil {
		return err
	}

	// if lvmsize equal/bigger than pv size, no do expand.
	if sizeInt >= pvSize {
		return nil
	}
	logging.GetLogger().Infof("NodeExpandVolume:: volumeId: %s, devicePath: %s, from size: %d, to Size: %d%s", volumeID, devicePath, sizeInt, pvSize, unit)

	// resize lvm volume
	// lvextend -L3G /dev/vgtest/lvm-5db74864-ea6b-11e9-a442-00163e07fb69
	resizeCmd := fmt.Sprintf("%s lvextend -L%d%s %s", types.NsenterCmd, pvSize, unit, devicePath)
	_, err = utils.Run(resizeCmd)
	if err != nil {
		return err
	}

	// use resizer to expand volume filesystem
	realExec := k8sexec.New()
	resizer := resizefs.NewResizeFs(&k8smount.SafeFormatAndMount{Interface: ns.k8smounter, Exec: realExec})
	ok, err := resizer.Resize(devicePath, targetPath)
	if err != nil {
		logging.GetLogger().Errorf("NodeExpandVolume:: Resize Error, volumeId: %s, devicePath: %s, volumePath: %s, err: %s", volumeID, devicePath, targetPath, err.Error())
		return err
	}
	if !ok {
		logging.GetLogger().Errorf("NodeExpandVolume:: Resize failed, volumeId: %s, devicePath: %s, volumePath: %s", volumeID, devicePath, targetPath)
		return status.Error(codes.Internal, "Fail to resize volume fs")
	}
	logging.GetLogger().Infof("NodeExpandVolume:: resizefs successful volumeId: %s, devicePath: %s, volumePath: %s", volumeID, devicePath, targetPath)
	return nil
}

func (ns *nodeServer) getPvSize(volumeID string) (int64, string) {
	pv, err := ns.client.CoreV1().PersistentVolumes().Get(volumeID, metav1.GetOptions{})
	if err != nil {
		logging.GetLogger().Errorf("lvcreate: fail to get pv, err: %v", err)
		return 0, ""
	}
	pvQuantity := pv.Spec.Capacity["storage"]
	pvSize := pvQuantity.Value()
	pvSizeGB := pvSize / (1024 * 1024 * 1024)

	if pvSizeGB == 0 {
		pvSizeMB := pvSize / (1024 * 1024)
		return pvSizeMB, "m"
	}
	return pvSizeGB, "g"
}

// delete lvm volume
func (ns *nodeServer) deleteVolume(ctx context.Context, devicePath string) error {
	cmd := fmt.Sprintf("%s lvremove -f %s ", types.NsenterCmd, devicePath)
	_, err := utils.Run(cmd)
	if err != nil {
		logging.GetLogger().Errorf("Delete LVM volume fail, err:%v", err.Error())
		return err
	}

	logging.GetLogger().Infof("Successful delete LVM volume: %s", devicePath)

	return nil
}
