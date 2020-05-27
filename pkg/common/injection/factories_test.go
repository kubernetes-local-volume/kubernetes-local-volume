package injection

import (
	"context"
	"testing"
)

func injectFooFactory(ctx context.Context) context.Context {
	return ctx
}

func injectBarFactory(ctx context.Context) context.Context {
	return ctx
}

func TestRegisterInformerFactory(t *testing.T) {
	i := &impl{}

	if want, got := 0, len(i.GetInformerFactories()); got != want {
		t.Errorf("GetInformerFactories() = %d, wanted %d", want, got)
	}

	i.RegisterInformerFactory(injectFooFactory)

	if want, got := 1, len(i.GetInformerFactories()); got != want {
		t.Errorf("GetInformerFactories() = %d, wanted %d", want, got)
	}

	i.RegisterInformerFactory(injectBarFactory)

	if want, got := 2, len(i.GetInformerFactories()); got != want {
		t.Errorf("GetInformerFactories() = %d, wanted %d", want, got)
	}
}
