package injection

import (
	"context"

	"k8s.io/client-go/rest"
)

// ClientInjector holds the type of a callback that attaches a particular
// client type to a context.
type ClientInjector func(context.Context, *rest.Config) context.Context

func (i *impl) RegisterClient(ci ClientInjector) {
	i.m.Lock()
	defer i.m.Unlock()

	i.clients = append(i.clients, ci)
}

func (i *impl) GetClients() []ClientInjector {
	i.m.RLock()
	defer i.m.RUnlock()

	// Copy the slice before returning.
	return append(i.clients[:0:0], i.clients...)
}
