package gc

import (
	"context"
	"flag"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/client"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/informers/core/v1/persistentvolume"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	internaltypes "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
)

var (
	nodeID = flag.String("nodeid", "", "node id")
)

func NewGC(
	ctx context.Context,
) *controller.Impl {
	flag.Parse()
	logger := logging.FromContext(ctx)
	client := client.Get(ctx)
	pvInformer := persistentvolume.Get(ctx)

	r := &Reconciler{
		nodeID:     *nodeID,
		client:     client,
		pvInformer: pvInformer,
		pvLister:   pvInformer.Lister(),
	}

	impl := controller.NewImpl(r, logger, ReconcilerName)

	pvInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filter(*nodeID),
		Handler:    controller.HandleAll(impl.Enqueue),
	})

	logger.Info("GC Started")
	return impl
}

func filter(nodeID string) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		pv, ok := obj.(*v1.PersistentVolume)
		if !ok {
			return false
		}

		return internaltypes.IsPVInMyNode(pv, nodeID)
	}
}
