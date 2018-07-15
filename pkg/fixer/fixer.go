package fixer

import (
	"context"
	"math/rand"
	"time"

	"github.com/juan-lee/ahabd/pkg/fixer/stats"
	"github.com/weaveworks/kured/pkg/delaytick"
)

// Fixer is the interface implemented for anything that needs assessment and
// repair.
type Fixer interface {
	NeedsFixing(ctx context.Context) bool
	Fix(ctx context.Context) error
	Stats() stats.Stats
}

// Fix checks the health of a resource and performs repairs if it's unhealthy.
// Metrics are kept for the operation.
func Fix(ctx context.Context, f Fixer) {
	if f.NeedsFixing(ctx) {
		s := f.Stats()
		if s == nil {
			s = &stats.NullStats{}
		}
		s.IncNeedsFixing()
		if err := f.Fix(ctx); err != nil {
			s.IncFixFail()
		}
		s.IncFixed()
	}
}

// PeriodicFix performs a Fix at the specified interval.
func PeriodicFix(ctx context.Context, f Fixer, period time.Duration) error {
	var err error
	tick := delaytick.New(rand.NewSource(time.Now().UnixNano()), period)
	for _ = range tick {
		if err = ctx.Err(); err != nil {
			return err
		}
		Fix(ctx, f)
	}

	return nil
}
