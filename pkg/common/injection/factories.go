package injection

import (
	"context"
)

// InformerFactoryInjector holds the type of a callback that attaches a particular
// factory type to a context.
type InformerFactoryInjector func(context.Context) context.Context

func (i *impl) RegisterInformerFactory(ifi InformerFactoryInjector) {
	i.m.Lock()
	defer i.m.Unlock()

	i.factories = append(i.factories, ifi)
}

func (i *impl) GetInformerFactories() []InformerFactoryInjector {
	i.m.RLock()
	defer i.m.RUnlock()

	// Copy the slice before returning.
	return append(i.factories[:0:0], i.factories...)
}
