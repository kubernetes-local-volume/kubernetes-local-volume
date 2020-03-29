package scheduler

import (
	"context"

	corev1 "k8s.io/client-go/listers/core/v1"
	storagev1 "k8s.io/client-go/listers/storage/v1"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/injection/informers/storage/v1alpha1/localvolume"
	pvc "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/informers/core/v1/persistentvolumeclaim"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/informers/core/v1/pod"
	sc "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/informers/storage/v1/storageclass"
	lv "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/listers/storage/v1alpha1"
)

type LocalVolumeScheduler struct {
	podLister          corev1.PodLister
	pvcLister          corev1.PersistentVolumeClaimLister
	storageclassLister storagev1.StorageClassLister
	localvolumeLister  lv.LocalVolumeLister
	ctx                context.Context
}

func NewLocalVolumeScheduler(ctx context.Context) *LocalVolumeScheduler {
	podInformer := pod.Get(ctx)
	pvcInformer := pvc.Get(ctx)
	scInformer := sc.Get(ctx)
	lvInformer := localvolume.Get(ctx)

	return &LocalVolumeScheduler{
		podLister:          podInformer.Lister(),
		pvcLister:          pvcInformer.Lister(),
		storageclassLister: scInformer.Lister(),
		localvolumeLister:  lvInformer.Lister(),
		ctx:                ctx,
	}
}
