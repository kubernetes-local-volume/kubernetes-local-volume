package injection

import (
	"context"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"

	"k8s.io/client-go/rest"
)

// InformerInjector holds the type of a callback that attaches a particular
// informer type to a context.
type InformerInjector func(context.Context) (context.Context, controller.Informer)

func (i *impl) RegisterInformer(ii InformerInjector) {
	i.m.Lock()
	defer i.m.Unlock()

	i.informers = append(i.informers, ii)
}

func (i *impl) GetInformers() []InformerInjector {
	i.m.RLock()
	defer i.m.RUnlock()

	// Copy the slice before returning.
	return append(i.informers[:0:0], i.informers...)
}

func (i *impl) SetupInformers(ctx context.Context, cfg *rest.Config) (context.Context, []controller.Informer) {
	// Based on the reconcilers we have linked, build up a set of clients and inject
	// them onto the context.
	for _, ci := range i.GetClients() {
		ctx = ci(ctx, cfg)
	}

	// Based on the reconcilers we have linked, build up a set of informer factories
	// and inject them onto the context.
	for _, ifi := range i.GetInformerFactories() {
		ctx = ifi(ctx)
	}

	// Based on the reconcilers we have linked, build up a set of duck informer factories
	// and inject them onto the context.
	for _, duck := range i.GetDucks() {
		ctx = duck(ctx)
	}

	// Based on the reconcilers we have linked, build up a set of informers
	// and inject them onto the context.
	var inf controller.Informer
	informers := make([]controller.Informer, 0, len(i.GetInformers()))
	for _, ii := range i.GetInformers() {
		ctx, inf = ii(ctx)
		informers = append(informers, inf)
	}
	return ctx, informers
}
