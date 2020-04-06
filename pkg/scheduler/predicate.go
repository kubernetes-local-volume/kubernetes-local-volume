package scheduler

import (
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"
)

func (lvs *LocalVolumeScheduler) PredicateHandler(args schedulerapi.ExtenderArgs) *schedulerapi.ExtenderFilterResult {
	pod := args.Pod
	canSchedule := make([]v1.Node, 0, len(args.Nodes.Items))
	canScheduleNodeNames := make(map[string]string)
	canNotSchedule := make(map[string]string)
	logger := logging.FromContext(lvs.ctx)

	for _, node := range args.Nodes.Items {
		result, err := lvs.predicate(*pod, node)
		if err != nil {
			canNotSchedule[node.Name] = err.Error()
		} else if result {
			canSchedule = append(canSchedule, node)
			canScheduleNodeNames[node.Name] = ""
		}
	}

	result := schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: canSchedule,
		},
		FailedNodes: canNotSchedule,
		Error:       "",
	}

	logger.Infof("local volume scheduler handle predicate: pod(%s) namespace(%s) can schedule nodes(%v)",
		pod.Name, pod.Namespace, canScheduleNodeNames)

	return &result
}

func (lvs *LocalVolumeScheduler) predicate(pod v1.Pod, node v1.Node) (bool, error) {
	logger := logging.FromContext(lvs.ctx)
	requestSize := lvs.getPodLocalVolumeRequestSize(&pod)
	lvFreeSize := lvs.getNodeFreeSize(node.Name)

	logger.Infof("local volume scheduler handle predicate: pod(%s) namespace(%s) request size(%v), node(%s) free size(%v)",
		pod.Name, pod.Namespace, requestSize, lvFreeSize)

	if lvFreeSize > requestSize {
		return true, nil
	}
	return false, nil
}
