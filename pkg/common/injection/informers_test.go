package injection

import (
	"context"
	"testing"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"

	"k8s.io/client-go/rest"
)

type fakeInformer struct{}

// HasSynced implements controller.Informer
func (*fakeInformer) HasSynced() bool {
	return false
}

// Run implements controller.Informer
func (*fakeInformer) Run(<-chan struct{}) {}

var _ controller.Informer = (*fakeInformer)(nil)

func injectFooInformer(ctx context.Context) (context.Context, controller.Informer) {
	return ctx, nil
}

func injectBarInformer(ctx context.Context) (context.Context, controller.Informer) {
	return ctx, nil
}

func TestRegisterInformersAndSetup(t *testing.T) {
	i := &impl{}

	if want, got := 0, len(i.GetInformers()); got != want {
		t.Errorf("GetInformerFactories() = %d, wanted %d", want, got)
	}

	i.RegisterClient(injectFoo)
	i.RegisterClient(injectBar)

	i.RegisterInformerFactory(injectFooFactory)
	i.RegisterInformerFactory(injectBarFactory)

	i.RegisterInformer(injectFooInformer)
	i.RegisterInformer(injectBarInformer)

	_, infs := i.SetupInformers(context.Background(), &rest.Config{})

	if want, got := 2, len(infs); got != want {
		t.Errorf("SetupInformers() = %d, wanted %d", want, got)
	}
}
