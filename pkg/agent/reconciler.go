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
	nodeID     string
	client     versioned.Interface
	lvInformer v1alpha1.LocalVolumeInformer
	lvLister   nlvslisters.LocalVolumeLister
	pvLister   corev1.PersistentVolumeLister
}

func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorw("Invalid resource key", zap.Error(err))
		return nil
	}

	// not concern other node
	if name != r.nodeID {
		return nil
	}

	// Get NodeLocalVolumeStorage resource with this namespace/name
	original, err := r.lvLister.LocalVolumes(namespace).Get(name)
	n := original.DeepCopy()

	if err := r.reconciler(n); err != nil {
		return err
	}

	logger.Infof("Reconcile NodeLocalVolumeStorage Resource Name = %s, Namespace = %s", name, namespace)
	return nil
}

func (r *Reconciler) reconciler(lv *nlvsv1alpha1.LocalVolume) error {
	logger := logging.GetLogger()
	isNlvsChange := false
	vgInfo := lvm.GetVGInfo(types.VGName)
	if vgInfo == nil {
		logger.Infof("reconciler %s not get vg(%s)", lv.Name, types.VGName)
		return nil
	}

	// 1. update total size
	totalSize := uint64(math.Floor(vgInfo.VgSize / 1024))
	if totalSize != lv.Status.TotalSize {
		lv.Status.TotalSize = totalSize
		isNlvsChange = true
	}

	// 2. update free size
	freeSize := uint64(math.Floor(vgInfo.VgFree / 1024))
	if freeSize != lv.Status.FreeSize {
		lv.Status.FreeSize = freeSize
		isNlvsChange = true
	}

	// 3. update preallocated info
	myNodePVs := r.getMyNodeBoundedPV()
	for pvName, _ := range myNodePVs {
		if _, ok := lv.Status.PreAllocated[pvName]; ok {
			delete(lv.Status.PreAllocated, pvName)
			isNlvsChange = true
		}
	}

	// 4. update nlvs
	if isNlvsChange {
		_, err := r.client.LocalV1alpha1().LocalVolumes(lv.Namespace).UpdateStatus(lv)
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
