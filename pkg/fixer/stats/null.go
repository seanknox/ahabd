package stats

type NullStats struct{}

func (s *NullStats) IncNeedsFixing() {
}

func (s *NullStats) IncFixed() {
}

func (s *NullStats) IncFixFail() {
}
