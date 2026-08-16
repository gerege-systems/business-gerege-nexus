package main

import (
	"path/filepath"
	"testing"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/catalog"
	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"

	"github.com/gerege-systems/business-gerege-nexus/modules/billing"
	"github.com/gerege-systems/business-gerege-nexus/modules/contacts"
	"github.com/gerege-systems/business-gerege-nexus/modules/inventory"
	"github.com/gerege-systems/business-gerege-nexus/modules/products"
)

// The bundled catalogue has to agree with the binary it ships inside.
//
// The platform refuses to start on a catalogue whose apps disagree with the
// modules compiled into it — catalogue integrity is a boot failure, not a
// warning. So a disagreement here is not a stale file, it is an instance that
// does not come up, discovered on the host rather than in CI.
//
// It is easy to arrive at without touching the catalogue: bumping the core is a
// one-line change, and a release that renames or re-versions one of the
// platform apps this catalogue copies leaves the copy behind. That is exactly
// what backend/v1.7.0 did to organisation — 2.0.0 "Directory" became 3.0.0
// "Organisation" — and this is the test that would have said so.
//
// What it can check is the half that lives here: the four modules this
// repository compiles. The platform's own are checked at boot by the core,
// because their code is in its internal packages and cannot be imported.
func TestTheBundledCatalogueAgreesWithThisBinary(t *testing.T) {
	apps, err := catalog.LoadFile(filepath.Join("catalog", "apps.json"), "")
	if err != nil {
		t.Fatalf("the bundled catalogue does not load: %v", err)
	}
	if len(apps) == 0 {
		t.Fatal("the bundled catalogue is empty")
	}

	p := nexus.NewPlatform(nil, nil)
	contacts.New(p)
	products.New(p)
	inventory.New(p, false)
	billing.New(p)

	for _, app := range apps {
		module, compiled := nexus.Get(app.ID)
		if !compiled {
			// A platform app from the core, or an external one. Both are
			// legitimate entries this test cannot speak for.
			continue
		}
		if module.Version() != app.Version {
			t.Errorf("%s is compiled at %s and the catalogue says %s",
				app.ID, module.Version(), app.Version)
		}
		if module.Version() != app.Manifest.Version {
			t.Errorf("%s is compiled at %s and its manifest says %s",
				app.ID, module.Version(), app.Manifest.Version)
		}
	}
}
