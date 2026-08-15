package main

import (
	"testing"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"

	"github.com/gerege-systems/commerce-gerege-nexus/modules/billing"
	"github.com/gerege-systems/commerce-gerege-nexus/modules/inventory"
	"github.com/gerege-systems/commerce-gerege-nexus/modules/products"
)

// What each module declares to nexus.AccessPolicy, asserted here because the
// platform stopped deciding it.
//
// It used to: two switch statements in the core listed every app by id and said
// which permission gated its menu and its routes. All three of these modules
// were in both. Had they moved before that changed, the switches would simply
// have stopped matching — no compile error, no failing test — and every route
// in this product would have become reachable by any member of a tenant that
// installed it. A permission that disappears during a move is the worst shape
// this bug takes, because from the outside the app still works.
//
// So the answers travel with the modules, and this is the assertion that they
// still say what we think. nil pointers: none of these methods touches a field.
func TestEachModuleStillDeclaresItsAccessPolicy(t *testing.T) {
	for _, tc := range []struct {
		name         string
		module       nexus.Module
		menu, prefix string
	}{
		{"products", (*products.Module)(nil), "products.read", "products"},
		{"inventory", (*inventory.Module)(nil), "inventory.read", "inventory"},
		{"billing", (*billing.BillingModule)(nil), "billing.read", "billing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := nexus.MenuPermissionOf(tc.module); got != tc.menu {
				t.Errorf("menu permission: got %q, want %q", got, tc.menu)
			}
			if got := nexus.RoutePermissionPrefixOf(tc.module); got != tc.prefix {
				t.Errorf("route prefix: got %q, want %q", got, tc.prefix)
			}
		})
	}
}

// The ids are what a tenant's installation rows, the catalogue manifests and
// every menu entry are keyed by. They did not change when the code moved
// repositories and must not: an id is the one thing about a module that a
// database remembers.
func TestTheModuleIdsSurvivedTheMove(t *testing.T) {
	for want, mod := range map[string]nexus.Module{
		"io.gerege.nexus.products":  (*products.Module)(nil),
		"io.gerege.nexus.inventory": (*inventory.Module)(nil),
		"io.gerege.nexus.billing":   (*billing.BillingModule)(nil),
	} {
		if got := mod.ID(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
