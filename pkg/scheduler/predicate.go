package scheduler

import (
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"
)

func (lvs *LocalVolumeScheduler) PredicateHandler(args schedulerapi.ExtenderArgs) *schedulerapi.ExtenderFilterResult {
	pod := args.Pod
	canSchedule := make([]v1.Node, 0, len(args.Nodes.Items))
	canNotSchedule := make(map[string]string)

	for _, node := range args.Nodes.Items {
		result, err := lvs.predicate(*pod, node)
		if err != nil {
			canNotSchedule[node.Name] = err.Error()
		} else if result {
			canSchedule = append(canSchedule, node)
		}
	}

	result := schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: canSchedule,
		},
		FailedNodes: canNotSchedule,
		Error:       "",
	}

	return &result
}

func (lvs *LocalVolumeScheduler) predicate(pod v1.Pod, node v1.Node) (bool, error) {
	requestSize := lvs.getPodLocalVolumeRequestSize(&pod)
	lv, err := lvs.localvolumeLister.LocalVolumes(v1.NamespaceDefault).Get(node.Name)
	if err != nil {
		return false, nil
	}
	lvFreeSize := lvs.getLocalVolumeStorageFreeSize(lv)
	if lvFreeSize > requestSize {
		return true, nil
	}
	return false, nil
}
