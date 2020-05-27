package injection

import (
	"context"
	"testing"

	"k8s.io/client-go/rest"
)

func injectFoo(ctx context.Context, cfg *rest.Config) context.Context {
	return ctx
}

func injectBar(ctx context.Context, cfg *rest.Config) context.Context {
	return ctx
}

func TestRegisterClient(t *testing.T) {
	i := &impl{}

	if want, got := 0, len(i.GetClients()); got != want {
		t.Errorf("GetClients() = %d, wanted %d", want, got)
	}

	i.RegisterClient(injectFoo)

	if want, got := 1, len(i.GetClients()); got != want {
		t.Errorf("GetClients() = %d, wanted %d", want, got)
	}

	i.RegisterClient(injectBar)

	if want, got := 2, len(i.GetClients()); got != want {
		t.Errorf("GetClients() = %d, wanted %d", want, got)
	}
}
