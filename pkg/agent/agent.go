package agent

import (
	"context"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/injection/informers/storage/v1alpha1/nodelocalvolumestorage"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
)

func NewAgent(
	ctx context.Context,
) *controller.Impl {
	logger := logging.FromContext(ctx)
	nlvsInformer := nodelocalvolumestorage.Get(ctx)

	c := &Reconciler{
		nlvsInformer: nlvsInformer,
		nlvsLister:   nlvsInformer.Lister(),
	}
	impl := controller.NewImpl(c, logger, ReconcilerName)

	nlvsInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	logger.Info("Agent Started")
	return impl
}
