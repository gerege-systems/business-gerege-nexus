package billing

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// invoicesCreated counts invoices issued, under the name the dashboards already
// use.
//
// It used to live in the platform's observability package, which meant the
// platform carried a counter named after this module's domain — and would have
// had to be edited the day this module moved to its own repository. The metric
// belongs where the event is known.
//
// Registered on the default registerer, which is what `/metrics` serves. A
// deployment that does not install billing does not export this series at all;
// an absent metric is a truer answer than a zero that never moves.
var invoicesCreated = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "invoices_created_total",
	Help: "Invoices issued across all tenants",
})

// Registered tolerantly rather than with MustRegister.
//
// During a split this module is compiled twice: once inside the platform it
// depends on, once in the distribution that adds it. Two registrations of one
// metric name is a panic at init with MustRegister — the process does not
// start, and the failure is a stack trace in a collector, nowhere near the
// cause. A module should not be able to stop a binary from booting because
// somebody compiled an older copy of it alongside.
//
// AlreadyRegisteredError means the other copy is counting and this one is not,
// which is a transitional half-truth and better than not booting. Any other
// error is a real problem with the metric itself and still panics.
func init() {
	if err := prometheus.Register(invoicesCreated); err != nil {
		var already prometheus.AlreadyRegisteredError
		if !errors.As(err, &already) {
			panic(err)
		}
	}
}
