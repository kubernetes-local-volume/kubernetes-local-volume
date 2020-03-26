package agent

import (
	"context"
	"math"

	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	nlvsv1alpha1 "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/apis/storage/v1alpha1"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/clientset/versioned"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/informers/externalversions/storage/v1alpha1"
	nlvslisters "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/listers/storage/v1alpha1"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/lvm"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
)

const (
	// ReconcilerName is the name of the reconciler
	ReconcilerName = "agent"
)

type Reconciler struct {
	nodeID       string
	client       versioned.Interface
	nlvsInformer v1alpha1.NodeLocalVolumeStorageInformer
	nlvsLister   nlvslisters.NodeLocalVolumeStorageLister
	pvLister     corev1.PersistentVolumeLister
}

func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorw("Invalid resource key", zap.Error(err))
		return nil
	}

	// Get NodeLocalVolumeStorage resource with this namespace/name
	original, err := r.nlvsLister.NodeLocalVolumeStorages(namespace).Get(name)
	nlvs := original.DeepCopy()

	if err := r.reconciler(nlvs); err != nil {
		return err
	}

	logger.Infof("Reconcile NodeLocalVolumeStorage Resource Name = %s, Namespace = %s", name, namespace)
	return nil
}

func (r *Reconciler) reconciler(nlvs *nlvsv1alpha1.NodeLocalVolumeStorage) error {
	isNlvsChange := false

	// 1. update total size
	totalSize, err := lvm.VGTotalSize(types.VGName)
	if err == nil {
		totalSize = uint64(math.Floor(float64(totalSize / 1024)))
		if totalSize != nlvs.Status.TotalSize {
			nlvs.Status.TotalSize = totalSize
			isNlvsChange = true
		}
	}

	// 2. update used size
	if freeSize, err := lvm.VGFreeSize(types.VGName); err == nil {
		usedSize := uint64(math.Floor(float64(totalSize-freeSize) / 1024))
		if usedSize != nlvs.Status.UsedSize {
			nlvs.Status.UsedSize = usedSize
			isNlvsChange = true
		}
	}

	// 3. update preallocated info
	myNodePVs := r.getMyNodeBoundedPV()
	for pvName, _ := range myNodePVs {
		if _, ok := nlvs.Status.PreAllocated[pvName]; ok {
			delete(nlvs.Status.PreAllocated, pvName)
			isNlvsChange = true
		}
	}

	// 4. update nlvs
	if isNlvsChange {
		_, err := r.client.LocalV1alpha1().NodeLocalVolumeStorages(nlvs.Namespace).Update(nlvs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) getMyNodeBoundedPV() map[string]*v1.PersistentVolume {
	result := make(map[string]*v1.PersistentVolume)

	allPV, err := r.pvLister.List(labels.Everything())
	if err != nil {
		return result
	}

	for _, pv := range allPV {
		if isPVInMyNode(pv, r.nodeID) && pv.Status.Phase == v1.VolumeBound {
			result[pv.Name] = pv
		}
	}

	return result
}
