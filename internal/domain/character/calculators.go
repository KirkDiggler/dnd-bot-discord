package character

// ACCalculator defines the interface for calculating armor class
// This allows different rulesets to implement their own AC calculation logic
type ACCalculator interface {
	// Calculate computes the armor class for a character
	// The implementation should consider armor, shields, abilities, and any special rules
	Calculate(char *Character) int
}
