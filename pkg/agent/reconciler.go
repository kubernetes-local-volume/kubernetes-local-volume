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
	nlvsInformer v1alpha1.NodeInfoInformer
	nlvsLister   nlvslisters.NodeInfoLister
	pvLister     corev1.PersistentVolumeLister
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
	original, err := r.nlvsLister.NodeInfos(namespace).Get(name)
	n := original.DeepCopy()

	if err := r.reconciler(n); err != nil {
		return err
	}

	logger.Infof("Reconcile NodeLocalVolumeStorage Resource Name = %s, Namespace = %s", name, namespace)
	return nil
}

func (r *Reconciler) reconciler(n *nlvsv1alpha1.NodeInfo) error {
	logger := logging.GetLogger()
	isNlvsChange := false
	vgInfo := lvm.GetVGInfo(types.VGName)
	if vgInfo == nil {
		logger.Infof("reconciler %s not get vg(%s)", n.Name, types.VGName)
		return nil
	}

	// 1. update total size
	total := uint64(math.Floor(vgInfo.VgSize / 1024))
	logger.Infof("11111:%d, %d", total, n.Status.TotalSize)
	if total != n.Status.TotalSize {
		n.Status.TotalSize = total
		isNlvsChange = true
	}

	// 2. update used size
	usedSize := uint64(math.Floor((vgInfo.VgSize - vgInfo.VgFree) / 1024))
	logger.Infof("22222:%d, %d", usedSize, n.Status.UsedSize)
	if usedSize != n.Status.UsedSize {
		n.Status.UsedSize = usedSize
		isNlvsChange = true
	}

	// 3. update preallocated info
	myNodePVs := r.getMyNodeBoundedPV()
	for pvName, _ := range myNodePVs {
		if _, ok := n.Status.PreAllocated[pvName]; ok {
			delete(n.Status.PreAllocated, pvName)
			isNlvsChange = true
		}
	}

	logger.Infof("agent reconciler total(%d) used(%d) isChange(%v) %d %d",
		n.Status.TotalSize, n.Status.UsedSize, isNlvsChange, total, usedSize)

	// 4. update nlvs
	if isNlvsChange {
		_, err := r.client.LocalV1alpha1().NodeInfos(n.Namespace).UpdateStatus(n)
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
