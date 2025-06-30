package dnd5e

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"log"
	"strconv"
	"strings"

	"github.com/fadedpez/dnd5e-api/clients/dnd5e"

	apiEntities "github.com/fadedpez/dnd5e-api/entities"
)

func apiReferenceItemToClass(apiClass *apiEntities.ReferenceItem) *rulebook.Class {
	return &rulebook.Class{
		Key:  apiClass.Key,
		Name: apiClass.Name,
	}
}

func apiReferenceItemsToClasses(input []*apiEntities.ReferenceItem) []*rulebook.Class {
	output := make([]*rulebook.Class, len(input))
	for i, apiClass := range input {
		output[i] = apiReferenceItemToClass(apiClass)
	}
	return output
}

func apiReferenceItemToRace(input *apiEntities.ReferenceItem) *rulebook.Race {
	return &rulebook.Race{
		Key:  input.Key,
		Name: input.Name,
	}
}

func apiReferenceItemsToRaces(input []*apiEntities.ReferenceItem) []*rulebook.Race {
	output := make([]*rulebook.Race, len(input))
	for i, apiRace := range input {
		output[i] = apiReferenceItemToRace(apiRace)
	}

	return output
}

func apiRaceToRace(input *apiEntities.Race) *rulebook.Race {
	return &rulebook.Race{
		Key:                        input.Key,
		Name:                       input.Name,
		Speed:                      input.Speed,
		StartingProficiencyOptions: apiChoiceOptionToChoice(input.StartingProficiencyOptions),
		StartingProficiencies:      apiReferenceItemsToReferenceItems(input.StartingProficiencies),
		AbilityBonuses:             apiAbilityBonusesToAbilityBonuses(input.AbilityBonuses),
	}
}

func apiAbilityBonusesToAbilityBonuses(input []*apiEntities.AbilityBonus) []*character.AbilityBonus {
	output := make([]*character.AbilityBonus, len(input))
	for i, apiAbilityBonus := range input {
		output[i] = apiAbilityBonusToAbilityBonus(apiAbilityBonus)
	}

	return output
}

func apiAbilityBonusToAbilityBonus(input *apiEntities.AbilityBonus) *character.AbilityBonus {
	if input == nil {
		return nil
	}
	if input.AbilityScore == nil {
		return nil
	}

	return &character.AbilityBonus{
		Attribute: referenceItemKeyToAttribute(input.AbilityScore.Key),
		Bonus:     input.Bonus,
	}
}

func referenceItemKeyToAttribute(input string) shared.Attribute {
	switch input {
	case "str":
		return shared.AttributeStrength
	case "dex":
		return shared.AttributeDexterity
	case "con":
		return shared.AttributeConstitution
	case "int":
		return shared.AttributeIntelligence
	case "wis":
		return shared.AttributeWisdom
	case "cha":
		return shared.AttributeCharisma
	default:
		log.Fatalf("Unknown attribute %s", input)
		return shared.AttributeNone
	}
}

func apiProficiencyToProficiency(input *apiEntities.Proficiency) *rulebook.Proficiency {
	return &rulebook.Proficiency{
		Key:  input.Key,
		Name: input.Name,
		Type: apiProficiencyTypeToProficiencyType(input.Type),
	}
}

func apiProficiencyTypeToProficiencyType(input apiEntities.ProficiencyType) rulebook.ProficiencyType {
	switch input {
	case apiEntities.ProficiencyTypeArmor:
		return rulebook.ProficiencyTypeArmor
	case apiEntities.ProficiencyTypeWeapon:
		return rulebook.ProficiencyTypeWeapon
	case apiEntities.ProficiencyTypeTool:
		return rulebook.ProficiencyTypeTool
	case apiEntities.ProficiencyTypeSavingThrow:
		return rulebook.ProficiencyTypeSavingThrow
	case apiEntities.ProficiencyTypeSkill:
		return rulebook.ProficiencyTypeSkill
	case apiEntities.ProficiencyTypeInstrument:
		return rulebook.ProficiencyTypeInstrument
	default:
		// Silently handle unknown proficiency types
		return rulebook.ProficiencyTypeUnknown

	}
}
func apiClassToClass(input *apiEntities.Class) *rulebook.Class {
	return &rulebook.Class{
		Key:                      input.Key,
		Name:                     input.Name,
		HitDie:                   input.HitDie,
		ProficiencyChoices:       apiChoicesToChoices(input.ProficiencyChoices),
		Proficiencies:            apiReferenceItemsToReferenceItems(input.Proficiencies),
		StartingEquipmentChoices: apiChoicesToChoices(input.StartingEquipmentOptions),
		StartingEquipment:        apiStartingEquipmentsToStartingEquipments(input.StartingEquipment),
	}
}

func apiStartingEquipmentToStartingEquipment(input *apiEntities.StartingEquipment) *rulebook.StartingEquipment {
	return &rulebook.StartingEquipment{
		Quantity:  input.Quantity,
		Equipment: apiReferenceItemToReferenceItem(input.Equipment),
	}
}

func apiStartingEquipmentsToStartingEquipments(input []*apiEntities.StartingEquipment) []*rulebook.StartingEquipment {
	output := make([]*rulebook.StartingEquipment, len(input))
	for i, apiStartingEquipment := range input {
		output[i] = apiStartingEquipmentToStartingEquipment(apiStartingEquipment)
	}

	return output
}
func apiEquipmentInterfaceToEquipment(input dnd5e.EquipmentInterface) equipment.Equipment {
	if input == nil {
		return nil
	}

	switch equip := input.(type) {
	case *apiEntities.Equipment:
		return apiEquipmentToEquipment(equip)
	case *apiEntities.Weapon:
		return apiWeaponToWeapon(equip)
	case *apiEntities.Armor:
		return apiArmorToArmor(equip)
	default:
		// Silently handle unknown equipment types
		return nil
	}
}

func apiWeaponToWeapon(input *apiEntities.Weapon) *equipment.Weapon {
	return &equipment.Weapon{
		Base: equipment.BasicEquipment{
			Key:    input.Key,
			Name:   input.Name,
			Weight: input.Weight,
			Cost:   apiCostToCost(input.Cost),
		},
		WeaponCategory:  strings.ToLower(input.WeaponCategory), // Normalize to lowercase
		WeaponRange:     input.WeaponRange,
		CategoryRange:   input.CategoryRange,
		Properties:      apiReferenceItemsToReferenceItems(input.Properties),
		Damage:          apiDamageToDamage(input.Damage),
		TwoHandedDamage: apiDamageToDamage(input.TwoHandedDamage),
	}
}

func apiDamageToDamage(input *apiEntities.Damage) *damage.Damage {
	if input == nil {
		return nil
	}

	diceParts := strings.Split(input.DamageDice, "d")
	if len(diceParts) != 2 {
		log.Printf("Unknown dice format %s", input.DamageDice)
		return nil
	}

	diceCount, err := strconv.Atoi(diceParts[0])
	if err != nil {
		log.Printf("Unknown dice format %s", input.DamageDice)
		return nil
	}

	diceValue, err := strconv.Atoi(diceParts[1])
	if err != nil {
		log.Printf("Unknown dice format %s", input.DamageDice)
		return nil
	}

	return &damage.Damage{
		DiceCount:  diceCount,
		DiceSize:   diceValue,
		DamageType: apiDamageTypeToDamageType(input.DamageType),
	}
}

func apiDamageTypeToDamageType(input *apiEntities.ReferenceItem) damage.Type {
	if input == nil {
		return damage.TypeNone
	}

	switch input.Key {
	case "acid":
		return damage.TypeAcid
	case "bludgeoning":
		return damage.TypeBludgeoning
	case "cold":
		return damage.TypeCold
	case "fire":
		return damage.TypeFire
	case "force":
		return damage.TypeForce
	case "lightning":
		return damage.TypeLightning
	case "necrotic":
		return damage.TypeNecrotic
	case "piercing":
		return damage.TypePiercing
	case "poison":
		return damage.TypePoison
	case "psychic":
		return damage.TypePsychic
	case "radiant":
		return damage.TypeRadiant
	case "slashing":
		return damage.TypeSlashing
	case "thunder":
		return damage.TypeThunder
	default:
		// Silently handle unknown damage types
		return damage.TypeNone
	}
}

func apiArmorToArmor(input *apiEntities.Armor) *equipment.Armor {
	// Determine armor category from the API data
	var category equipment.ArmorCategory
	switch strings.ToLower(input.ArmorCategory) {
	case "light":
		category = equipment.ArmorCategoryLight
	case "medium":
		category = equipment.ArmorCategoryMedium
	case "heavy":
		category = equipment.ArmorCategoryHeavy
	case "shield":
		category = equipment.ArmorCategoryShield
	default:
		category = equipment.ArmorCategoryUnknown
	}

	return &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:    input.Key,
			Name:   input.Name,
			Weight: input.Weight,
			Cost:   apiCostToCost(input.Cost),
		},
		ArmorCategory: category,
		ArmorClass: &equipment.ArmorClass{
			Base:     input.ArmorClass.Base,
			DexBonus: input.ArmorClass.DexBonus,
			// MaxBonus is not available from the API, so medium armor limits
			// are handled in the AC calculator fallback logic
		},
		StealthDisadvantage: input.StealthDisadvantage,
	}
}

func apiEquipmentToEquipment(input *apiEntities.Equipment) *equipment.BasicEquipment {
	return &equipment.BasicEquipment{
		Key:    input.Key,
		Name:   input.Name,
		Weight: input.Weight,
		Cost:   apiCostToCost(input.Cost),
	}
}

func apiCostToCost(input *apiEntities.Cost) *shared.Cost {
	return &shared.Cost{
		Quantity: input.Quantity,
		Unit:     input.Unit,
	}
}

func apiChoicesToChoices(input []*apiEntities.ChoiceOption) []*shared.Choice {
	output := make([]*shared.Choice, len(input))
	for i, apiChoice := range input {
		output[i] = apiChoiceOptionToChoice(apiChoice)
	}

	return output
}

func apiChoiceOptionToChoice(input *apiEntities.ChoiceOption) *shared.Choice {
	if input == nil {
		return nil
	}

	if input.OptionList == nil {
		return nil
	}

	output := make([]shared.Option, len(input.OptionList.Options))

	for i, apiProficiency := range input.OptionList.Options {
		output[i] = apiOptionToOption(apiProficiency)
	}

	// Choice created successfully

	return &shared.Choice{
		Count:   input.ChoiceCount,
		Name:    input.Description,
		Type:    apiChoiceTypeToChoiceType(input.ChoiceType),
		Key:     "choice",
		Options: output,
	}
}

func apiChoiceTypeToChoiceType(input string) shared.ChoiceType {
	switch input {
	case "proficiencies":
		return shared.ChoiceTypeProficiency
	case "equipment":
		return shared.ChoiceTypeEquipment
	case "languages":
		return shared.ChoiceTypeLanguage
	default:
		// Silently handle unknown choice types
		return shared.ChoiceTypeUnset
	}
}

func apiOptionToOption(input apiEntities.Option) shared.Option {
	switch input.GetOptionType() {
	case apiEntities.OptionTypeReference:
		item, ok := input.(*apiEntities.ReferenceOption)
		if !ok || item.Reference == nil {
			return nil
		}

		return &shared.ReferenceOption{
			Reference: apiReferenceItemToReferenceItem(item.Reference),
		}
	case apiEntities.OptionTypeChoice:
		item, ok := input.(*apiEntities.ChoiceOption)
		if !ok {
			return nil
		}

		return apiChoiceOptionToChoice(item)
	case apiEntities.OptionalTypeCountedReference:
		item, ok := input.(*apiEntities.CountedReferenceOption)
		if !ok || item.Reference == nil {
			return nil
		}

		return &shared.CountedReferenceOption{
			Count:     item.Count,
			Reference: apiReferenceItemToReferenceItem(item.Reference),
		}
	case apiEntities.OptionTypeMultiple:
		item, ok := input.(*apiEntities.MultipleOption)
		if !ok || item.Items == nil {
			return nil
		}

		options := make([]shared.Option, len(item.Items))
		for i, apiOption := range item.Items {
			options[i] = apiOptionToOption(apiOption)
		}

		return &shared.MultipleOption{
			Items: options,
		}
	default:
		// Silently handle unknown option types
		return nil
	}
}

func apiReferenceItemsToReferenceItems(input []*apiEntities.ReferenceItem) []*shared.ReferenceItem {
	output := make([]*shared.ReferenceItem, len(input))
	for i, apiReferenceItem := range input {
		output[i] = apiReferenceItemToReferenceItem(apiReferenceItem)
	}

	return output
}
func apiReferenceItemToReferenceItem(input *apiEntities.ReferenceItem) *shared.ReferenceItem {
	return &shared.ReferenceItem{
		Key:  input.Key,
		Name: input.Name,
		Type: typeStringToReferenceType(input.Type),
	}
}

func typeStringToReferenceType(input string) shared.ReferenceType {
	switch input {
	case "equipment":
		return shared.ReferenceTypeEquipment
	case "proficiencies":
		return shared.ReferenceTypeProficiency
	case "languages":
		return shared.ReferenceTypeLanguage
	case "ability-scores":
		return shared.ReferenceTypeAbilityScore
	case "skills":
		return shared.ReferenceTypeSkill
	case "weapon-properties":
		return shared.ReferenceTypeWeaponProperty
	default:
		// Don't log for known but unhandled types
		if input != "2014" {
			log.Println("Unknown reference type: ", input)
		}
		return shared.ReferenceTypeUnset
	}
}
