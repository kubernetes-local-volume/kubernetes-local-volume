package agent

import (
	"context"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/client"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/informers/core/v1/persistentvolume"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	internaltypes "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
)

func NewGC(
	ctx context.Context,
) *controller.Impl {
	logger := logging.FromContext(ctx)
	client := client.Get(ctx)
	pvInformer := persistentvolume.Get(ctx)

	r := &GCReconciler{
		nodeID:     *nodeID,
		client:     client,
		pvInformer: pvInformer,
		pvLister:   pvInformer.Lister(),
	}

	impl := controller.NewImpl(r, logger, GCReconcilerName)

	pvInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: gcFilter(*nodeID),
		Handler:    controller.HandleAll(impl.Enqueue),
	})

	logger.Info("GC Started")
	return impl
}

func gcFilter(nodeID string) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		pv, ok := obj.(*v1.PersistentVolume)
		if !ok {
			return false
		}

		return internaltypes.IsPVInMyNode(pv, nodeID)
	}
}
