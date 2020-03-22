package agent

import (
	"context"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/informers/externalversions/storage/v1alpha1"
	listers "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/listers/storage/v1alpha1"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"

	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

const (
	// ReconcilerName is the name of the reconciler
	ReconcilerName = "agent"
)

type Reconciler struct {
	nlvsInformer v1alpha1.NodeLocalVolumeStorageInformer
	nlvsLister   listers.NodeLocalVolumeStorageLister
}

func (c *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorw("Invalid resource key", zap.Error(err))
		return nil
	}

	// Get NodeLocalVolumeStorage resource with this namespace/name
	_, err = c.nlvsLister.NodeLocalVolumeStorages(namespace).Get(name)

	logger.Infof("Reconcile NodeLocalVolumeStorage Resource Name = %s, Namespace = %s", name, namespace)
	return nil
}
