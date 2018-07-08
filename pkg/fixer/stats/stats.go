package stats

// Stats is the interface implemented to track Fixer metrics
type Stats interface {
	IncNeedsFixing()
	IncFixed()
	IncFixFail()
}
