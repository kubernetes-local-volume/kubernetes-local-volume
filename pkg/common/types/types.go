package types

import (
	"strings"

	"k8s.io/api/core/v1"
)

var (
	// DeviceChars is chars of a device
	DeviceChars = []string{"b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
)

const (
	// driver name
	DriverName = "local.volume.csi.kubernetes.io"
	// version
	Version = "0.0.1"
)

const (
	// TopologyNodeKey tag
	TopologyNodeKey = "topology.local.volume.csi/hostname"
	// VG Name
	VGName = "local-volume-csi"
	// NsenterCmd is the nsenter command
	NsenterCmd = "/nsenter --mount=/proc/1/ns/mnt"
)

const (
	// gc tag
	LocalVolumeGCTag = "finalizers.localvolume.kubernetes.io/gc"
)

func MakePVCKey(namespace, name string) string {
	return namespace + "/" + name
}

func SplitPVCKey(key string) (string, string) {
	parts := strings.Split(key, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func IsPVInMyNode(pv *v1.PersistentVolume, nodeID string) bool {
	if pv.Spec.NodeAffinity == nil {
		return false
	}
	if pv.Spec.NodeAffinity.Required == nil {
		return false
	}
	if pv.Spec.NodeAffinity.Required.NodeSelectorTerms == nil {
		return false
	}

	for _, match := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
		if match.MatchExpressions == nil {
			continue
		}
		for _, v := range match.MatchExpressions {
			if v.Key == TopologyNodeKey {
				for _, node := range v.Values {
					if node == nodeID {
						return true
					}
				}
			}
		}
	}
	return false
}
