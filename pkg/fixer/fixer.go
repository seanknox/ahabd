package fixer

import (
	"math/rand"
	"time"

	"github.com/juan-lee/ahabd/pkg/fixer/stats"
	"github.com/weaveworks/kured/pkg/delaytick"
)

// Fixer is the interface implemented for anything that needs assessment and
// repair.
type Fixer interface {
	NeedsFixing() bool
	Fix() error
	Stats() stats.Stats
}

// Fix checks the health of a resource and performs repairs if it's unhealthy.
// Metrics are kept for the operation.
func Fix(f Fixer) error {
	if f.NeedsFixing() {
		s := f.Stats()
		if s == nil {
			s = &stats.NullStats{}
		}
		s.IncNeedsFixing()
		if err := f.Fix(); err != nil {
			s.IncFixFail()
			return err
		}
		s.IncFixed()
	}

	return nil
}

// PeriodicFix performs a Fix at the specified interval.
func PeriodicFix(f Fixer, period time.Duration) {
	source := rand.NewSource(time.Now().UnixNano())
	tick := delaytick.New(source, period)
	for _ = range tick {
		Fix(f)
	}
}
