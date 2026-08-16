/*
 * Gerege Nexus — Commerce
 * Copyright (c) 2026 Gerege Systems Development Team, Gerege Nomadica Foundation.
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
	"log/slog"
	"os"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/platform"

	"github.com/gerege-systems/business-gerege-nexus/modules/billing"
	"github.com/gerege-systems/business-gerege-nexus/modules/contacts"
	"github.com/gerege-systems/business-gerege-nexus/modules/inventory"
	"github.com/gerege-systems/business-gerege-nexus/modules/products"
)

func main() {
	// The error is checked, and the exit code is the point: a distribution that
	// cannot start — a catalogue that disagrees with its modules, a database it
	// cannot reach — must not exit 0 and read as a clean shutdown to whatever is
	// supervising it. The App Store's main.go has done this since it was
	// written; this one dropped the value.
	err := platform.Run(platform.Options{
		Modules: func(p nexus.Platform) {
			// products first: inventory's stock rows reference it, and the
			// dependency between the modules is declared rather than implied
			// by this order — but constructing it first keeps the reading
			// order and the dependency order the same.
			// Who this business sells to and buys from. It was the
			// platform's until the platform stopped pretending everybody has
			// customers: departments and staff are universal, a contact
			// register is what a business keeps.
			contacts.New(p)
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
	if err != nil {
		slog.Error("commerce stopped", "error", err)
		os.Exit(1)
	}
}
