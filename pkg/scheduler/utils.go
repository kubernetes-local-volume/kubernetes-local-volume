package scheduler

import (
	"math"

	corev1 "k8s.io/api/core/v1"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
)

func (lvs *LocalVolumeScheduler) getPodLocalVolumeRequestSize(pod *corev1.Pod) uint64 {
	var result uint64

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcName := volume.PersistentVolumeClaim.ClaimName

			// get pvc
			pvc, err := lvs.pvcLister.PersistentVolumeClaims(pod.Namespace).Get(pvcName)
			if err != nil {
				continue
			}

			// get storageclass
			sc, err := lvs.storageClassLister.Get(*pvc.Spec.StorageClassName)
			if err != nil {
				continue
			}

			if types.DriverName == sc.Provisioner {
				size, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				if ok {
					realSize := uint64(math.Ceil(float64(size.Value()) / 1024 / 1024 / 1024))
					result = result + realSize
				}
			}
		}
	}
	return result
}

func (lvs *LocalVolumeScheduler) getPodLocalVolumePVCNames(pod *corev1.Pod) map[string]string {
	result := make(map[string]string)

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcName := volume.PersistentVolumeClaim.ClaimName

			// get pvc
			pvc, err := lvs.pvcLister.PersistentVolumeClaims(pod.Namespace).Get(pvcName)
			if err != nil {
				continue
			}

			// get storageclass
			sc, err := lvs.storageClassLister.Get(*pvc.Spec.StorageClassName)
			if err != nil {
				continue
			}

			if sc.Provisioner == types.DriverName {
				result[types.MakePVCKey(pvc.Namespace, pvc.Name)] = ""
			}
		}
	}
	return result
}

func (lvs *LocalVolumeScheduler) getNodeFreeSize(nodeName string) uint64 {
	lv, err := lvs.localVolumeLister.LocalVolumes(corev1.NamespaceDefault).Get(nodeName)
	if err != nil {
		return 0
	}

	var preallocateSize uint64
	for key := range lv.Status.PreAllocated {
		pvcNS, pvcName := types.SplitPVCKey(key)
		pvc, err := lvs.pvcLister.PersistentVolumeClaims(pvcNS).Get(pvcName)
		if err != nil {
			continue
		}

		size, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		if !ok {
			continue
		}
		realSize := uint64(math.Ceil(float64(size.Value()) / 1024 / 1024 / 1024))
		preallocateSize = preallocateSize + realSize
	}
	return lv.Status.FreeSize - preallocateSize
}
