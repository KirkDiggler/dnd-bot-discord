package shared

type Attribute string

var Attributes = []Attribute{AttributeStrength, AttributeDexterity, AttributeConstitution, AttributeIntelligence, AttributeWisdom, AttributeCharisma}

const (
	AttributeNone         Attribute = ""
	AttributeStrength     Attribute = "Str"
	AttributeDexterity    Attribute = "Dex"
	AttributeConstitution Attribute = "Con"
	AttributeIntelligence Attribute = "Int"
	AttributeWisdom       Attribute = "Wis"
	AttributeCharisma     Attribute = "Cha"
)
