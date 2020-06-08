package agent

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/utils"
)

const (
	// ReconcilerName is the name of the reconciler
	GCReconcilerName = "RecycleLocalVolume"
)

var (
	LVNotFoundString = "Failed to find logical volume"
)

type GCReconciler struct {
	nodeID     string
	client     kubernetes.Interface
	pvInformer v1.PersistentVolumeInformer
	pvLister   listerv1.PersistentVolumeLister
}

func (r *GCReconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorw("Invalid resource key", zap.Error(err))
		return nil
	}

	original, err := r.pvLister.Get(name)
	if err != nil && errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}
	pv := original.DeepCopy()

	if err := r.reconciler(pv); err != nil {
		return err
	}
	return nil
}

func (r *GCReconciler) reconciler(pv *corev1.PersistentVolume) error {
	logger := logging.GetLogger()

	if pv.Status.Phase == corev1.VolumeReleased &&
		pv.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete &&
		utils.SliceContainsString(pv.Finalizers, types.LocalVolumeGCTag) {

		if err := r.deleteVolume(pv); err == nil {
			pv.Finalizers = utils.SliceRemoveString(pv.Finalizers, types.LocalVolumeGCTag)
			if _, err := r.client.CoreV1().PersistentVolumes().Update(pv); err != nil {
				logger.Errorf("GC Controller update pv error : %+v", err)
				return err
			} else {
				logger.Infof("GC Controller delete %s success", pv.Name)
			}
		}
	}
	return nil
}

func (r *GCReconciler) deleteVolume(pv *corev1.PersistentVolume) error {
	logger := logging.GetLogger()
	devicePath := filepath.Join("/dev/", types.VGName, "/", pv.Name)

	cmd := fmt.Sprintf("%s lvremove -f %s ", types.NsenterCmd, devicePath)
	_, err := utils.Run(cmd)
	if err != nil {
		if strings.Contains(err.Error(), LVNotFoundString) {
			return nil
		}
		logger.Errorf("GC Controller Delete LVM volume fail, err:%v", err.Error())
		return err
	}

	logger.Infof("GC Controller delete LVM volume: %s, PV(%s) success", devicePath, pv.Name)

	return nil
}
