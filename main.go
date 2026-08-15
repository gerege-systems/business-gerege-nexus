/*
 * Gerege Nexus — Commerce
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

// Command commercenexus runs the Gerege Nexus platform with the commerce
// modules compiled in: products, inventory and billing.
//
// Nothing else belongs in this file. A distribution is the platform plus its
// own modules; logic written here instead of in a module is logic no other
// deployment can have and no test can reach — the ecosystem strategy, §5 rule 3.
package main

import (
	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/platform"

	"github.com/gerege-systems/commerce-gerege-nexus/modules/billing"
	"github.com/gerege-systems/commerce-gerege-nexus/modules/inventory"
	"github.com/gerege-systems/commerce-gerege-nexus/modules/products"
)

func main() {
	platform.Run(platform.Options{
		Modules: func(p nexus.Platform) {
			// products first: inventory's stock rows reference it, and the
			// dependency between the modules is declared rather than implied
			// by this order — but constructing it first keeps the reading
			// order and the dependency order the same.
			products.New(p)
			// false: stock may not go negative. It is the platform's own
			// setting today, carried over unchanged rather than quietly
			// relaxed by the move — a distribution inheriting a different
			// answer to "may a warehouse ship what it does not have" would be
			// a behaviour change disguised as a refactor.
			inventory.New(p, false)
			billing.New(p)
		},
	})
}
