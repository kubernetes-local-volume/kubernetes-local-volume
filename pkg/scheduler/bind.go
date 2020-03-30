package scheduler

import (
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	corev1 "k8s.io/api/core/v1"
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
	logger := logging.FromContext(lvs.ctx)

	pod, err := lvs.podLister.Pods(podNamespace).Get(podName)
	if err != nil {
		return err
	}
	pvcNames := lvs.getPodLocalVolumePVCNames(pod)

	lv, err := lvs.localvolumeLister.LocalVolumes(corev1.NamespaceDefault).Get(node)
	if err != nil {
		return err
	}

	copylv := lv.DeepCopy()
	for _, v := range pvcNames {
		copylv.Status.PreAllocated[v] = ""
	}
	if _, err := lvs.client.LocalV1alpha1().LocalVolumes(corev1.NamespaceDefault).UpdateStatus(copylv); err != nil {
		return err
	}

	logger.Infof("pod(%s) namespace(%s) bind node(%s) success", podName, podNamespace, node)
	return nil
}
