package scheduler

import (
	"k8s.io/apimachinery/pkg/types"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"
)

func (lvs *LocalVolumeScheduler) BindHandler(args schedulerapi.ExtenderBindingArgs) *schedulerapi.ExtenderBindingResult {
	err := lvs.bind(args.PodName, args.PodNamespace, args.PodUID, args.Node)

	return &schedulerapi.ExtenderBindingResult{
		Error: err.Error(),
	}
}

func (lvs *LocalVolumeScheduler) bind(podName string, podNamespace string, podUID types.UID, node string) error {
	return nil
}