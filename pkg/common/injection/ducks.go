package injection

import (
	"context"
)

// DuckFactoryInjector holds the type of a callback that attaches a particular
// duck-type informer factory to a context.
type DuckFactoryInjector func(context.Context) context.Context

func (i *impl) RegisterDuck(ii DuckFactoryInjector) {
	i.m.Lock()
	defer i.m.Unlock()

	i.ducks = append(i.ducks, ii)
}

func (i *impl) GetDucks() []DuckFactoryInjector {
	i.m.RLock()
	defer i.m.RUnlock()

	// Copy the slice before returning.
	return append(i.ducks[:0:0], i.ducks...)
}
