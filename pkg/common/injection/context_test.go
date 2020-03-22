package injection

import (
	"context"
	"testing"
)

func TestGetBaseline(t *testing.T) {
	ctx := context.Background()

	if HasNamespaceScope(ctx) {
		t.Error("HasNamespaceScope() = true, wanted false")
	}

	want := "this-is-the-best-ns-evar"
	ctx = WithNamespaceScope(ctx, want)

	if !HasNamespaceScope(ctx) {
		t.Error("HasNamespaceScope() = false, wanted true")
	}

	if got := GetNamespaceScope(ctx); got != want {
		t.Errorf("GetNamespaceScope() = %v, wanted %v", got, want)
	}
}
