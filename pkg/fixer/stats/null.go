package stats

type nullStats struct{}

// NewNullStats returns a null metrics provider implementation.
func NewNullStats() Stats {
	return &nullStats{}
}

func (s *nullStats) IncNeedsFixing() {}

func (s *nullStats) IncFixed() {}

func (s *nullStats) IncFixFail() {}
