package injection

import (
	"context"
)

// nsKey is the key that namespaces are associated with on
// contexts returned by WithNamespaceScope.
type nsKey struct{}

// WithNamespaceScope associates a namespace scoping with the
// provided context, which will scope the informers produced
// by the downstream informer factories.
func WithNamespaceScope(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, nsKey{}, namespace)
}

// HasNamespaceScope determines whether the provided context has
// been scoped to a particular namespace.
func HasNamespaceScope(ctx context.Context) bool {
	return GetNamespaceScope(ctx) != ""
}

// GetNamespaceScope accesses the namespace associated with the
// provided context.  This should be called when the injection
// logic is setting up shared informer factories.
func GetNamespaceScope(ctx context.Context) string {
	value := ctx.Value(nsKey{})
	if value == nil {
		return ""
	}
	return value.(string)
}
