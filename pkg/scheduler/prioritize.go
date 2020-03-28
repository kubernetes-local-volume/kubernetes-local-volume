package scheduler

import (
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"
)

func (lvs *LocalVolumeScheduler) PrioritizeHandler(args schedulerapi.ExtenderArgs) (*schedulerapi.HostPriorityList, error) {
	return lvs.prioritize(*args.Pod, args.Nodes.Items)
}

func (lvs *LocalVolumeScheduler) prioritize(pod v1.Pod, nodes []v1.Node) (*schedulerapi.HostPriorityList, error) {
	var priorityList schedulerapi.HostPriorityList
	priorityList = make([]schedulerapi.HostPriority, len(nodes))
	for i, node := range nodes {
		priorityList[i] = schedulerapi.HostPriority{
			Host:  node.Name,
			Score: 0,
		}
	}
	return &priorityList, nil
}