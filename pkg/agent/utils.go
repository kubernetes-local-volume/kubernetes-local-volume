package agent

import (
	v1 "k8s.io/api/core/v1"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
)

func isPVInMyNode(pv *v1.PersistentVolume, nodeID string) bool {
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
			if v.Key == types.TopologyNodeKey {
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