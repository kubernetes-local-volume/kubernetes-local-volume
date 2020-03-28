package scheduler

import (
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"
)

func (lvs *LocalVolumeScheduler) PreemptionHandler(
	args schedulerapi.ExtenderPreemptionArgs,
) *schedulerapi.ExtenderPreemptionResult {
	nodeNameToMetaVictims := lvs.preemption(*args.Pod, args.NodeNameToVictims, args.NodeNameToMetaVictims)

	return &schedulerapi.ExtenderPreemptionResult{
		NodeNameToMetaVictims: nodeNameToMetaVictims,
	}
}

func (lvs *LocalVolumeScheduler) preemption(
	pod v1.Pod,
	victims map[string]*schedulerapi.Victims,
	metaVictims map[string]*schedulerapi.MetaVictims) map[string]*schedulerapi.MetaVictims {
	result := make(map[string]*schedulerapi.MetaVictims)
	return result
}