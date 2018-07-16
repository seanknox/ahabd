package stats

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type prometheusStats struct {
	needsFixingCounter prometheus.Counter
	fixedCounter       prometheus.Counter
	fixFailCounter     prometheus.Counter
}

// NewDefault returns a default prometheus metrics provider implementation.
func NewDefault(source, name string) Stats {
	return &prometheusStats{
		needsFixingCounter: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem: source,
			Name:      fmt.Sprintf("%s_restarts", name),
			Help:      fmt.Sprintf("# of %s restarts due to an unhealthy state.", name),
		}),
		fixedCounter: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem: source,
			Name:      fmt.Sprintf("%s_fixed", name),
			Help:      fmt.Sprintf("# of times %s was fixed due to an unhealthy state.", name),
		}),
		fixFailCounter: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem: source,
			Name:      fmt.Sprintf("%s_fix_failed", name),
			Help:      fmt.Sprintf("# of times fixing %s failed.", name),
		}),
	}
}

func (s prometheusStats) IncNeedsFixing() {
	s.needsFixingCounter.Inc()
}

func (s prometheusStats) IncFixed() {
	s.fixedCounter.Inc()
}

func (s prometheusStats) IncFixFail() {
	s.fixFailCounter.Inc()
}
