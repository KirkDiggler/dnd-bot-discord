package equipment

// ACProvider is an interface for equipment that provides AC bonuses
type ACProvider interface {
	Equipment
	GetACBase() int
	UsesDexBonus() bool
	GetMaxDexBonus() int // 0 means no limit
}
