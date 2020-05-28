package scheduler

import (
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"
)

func (lvs *LocalVolumeScheduler) PrioritizeHandler(args schedulerapi.ExtenderArgs) (*schedulerapi.HostPriorityList, error) {
	return lvs.prioritize(*args.Pod, args.Nodes.Items)
}

func (lvs *LocalVolumeScheduler) prioritize(pod v1.Pod, nodes []v1.Node) (*schedulerapi.HostPriorityList, error) {
	logger := logging.FromContext(lvs.ctx)
	requestSize := lvs.getPodLocalVolumeRequestSize(&pod)

	var priorityList schedulerapi.HostPriorityList
	priorityList = make([]schedulerapi.HostPriority, len(nodes))
	for i, node := range nodes {
		freeSize := lvs.getNodeFreeSize(node.Name)
		logger.Infof("local volume scheduler handle pod(%s, namespace = %s) requestsize(%d) prioritize: node(%s) free size(%d)",
			pod.Namespace, pod.Name, requestSize, node.Name, freeSize)

		priorityList[i] = schedulerapi.HostPriority{
			Host: node.Name,
		}

		if requestSize == 0 {
			priorityList[i].Score = 100
		} else if freeSize > requestSize {
			priorityList[i].Score = int64(freeSize) % 100
		} else {
			priorityList[i].Score = 0
		}
	}

	return &priorityList, nil
}
