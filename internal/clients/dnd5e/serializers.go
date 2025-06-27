package dnd5e

import (
	"log"
	"strconv"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"

	"github.com/fadedpez/dnd5e-api/clients/dnd5e"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	apiEntities "github.com/fadedpez/dnd5e-api/entities"
)

func apiReferenceItemToClass(apiClass *apiEntities.ReferenceItem) *entities.Class {
	return &entities.Class{
		Key:  apiClass.Key,
		Name: apiClass.Name,
	}
}

func apiReferenceItemsToClasses(input []*apiEntities.ReferenceItem) []*entities.Class {
	output := make([]*entities.Class, len(input))
	for i, apiClass := range input {
		output[i] = apiReferenceItemToClass(apiClass)
	}
	return output
}

func apiReferenceItemToRace(input *apiEntities.ReferenceItem) *entities.Race {
	return &entities.Race{
		Key:  input.Key,
		Name: input.Name,
	}
}

func apiReferenceItemsToRaces(input []*apiEntities.ReferenceItem) []*entities.Race {
	output := make([]*entities.Race, len(input))
	for i, apiRace := range input {
		output[i] = apiReferenceItemToRace(apiRace)
	}

	return output
}

func apiRaceToRace(input *apiEntities.Race) *entities.Race {
	return &entities.Race{
		Key:                        input.Key,
		Name:                       input.Name,
		Speed:                      input.Speed,
		StartingProficiencyOptions: apiChoiceOptionToChoice(input.StartingProficiencyOptions),
		StartingProficiencies:      apiReferenceItemsToReferenceItems(input.StartingProficiencies),
		AbilityBonuses:             apiAbilityBonusesToAbilityBonuses(input.AbilityBonuses),
	}
}

func apiAbilityBonusesToAbilityBonuses(input []*apiEntities.AbilityBonus) []*entities.AbilityBonus {
	output := make([]*entities.AbilityBonus, len(input))
	for i, apiAbilityBonus := range input {
		output[i] = apiAbilityBonusToAbilityBonus(apiAbilityBonus)
	}

	return output
}

func apiAbilityBonusToAbilityBonus(input *apiEntities.AbilityBonus) *entities.AbilityBonus {
	if input == nil {
		return nil
	}
	if input.AbilityScore == nil {
		return nil
	}

	return &entities.AbilityBonus{
		Attribute: referenceItemKeyToAttribute(input.AbilityScore.Key),
		Bonus:     input.Bonus,
	}
}

func referenceItemKeyToAttribute(input string) entities.Attribute {
	switch input {
	case "str":
		return entities.AttributeStrength
	case "dex":
		return entities.AttributeDexterity
	case "con":
		return entities.AttributeConstitution
	case "int":
		return entities.AttributeIntelligence
	case "wis":
		return entities.AttributeWisdom
	case "cha":
		return entities.AttributeCharisma
	default:
		log.Fatalf("Unknown attribute %s", input)
		return entities.AttributeNone
	}
}

func apiProficiencyToProficiency(input *apiEntities.Proficiency) *entities.Proficiency {
	return &entities.Proficiency{
		Key:  input.Key,
		Name: input.Name,
		Type: apiProficiencyTypeToProficiencyType(input.Type),
	}
}

func apiProficiencyTypeToProficiencyType(input apiEntities.ProficiencyType) entities.ProficiencyType {
	switch input {
	case apiEntities.ProficiencyTypeArmor:
		return entities.ProficiencyTypeArmor
	case apiEntities.ProficiencyTypeWeapon:
		return entities.ProficiencyTypeWeapon
	case apiEntities.ProficiencyTypeTool:
		return entities.ProficiencyTypeTool
	case apiEntities.ProficiencyTypeSavingThrow:
		return entities.ProficiencyTypeSavingThrow
	case apiEntities.ProficiencyTypeSkill:
		return entities.ProficiencyTypeSkill
	case apiEntities.ProficiencyTypeInstrument:
		return entities.ProficiencyTypeInstrument
	default:
		// Silently handle unknown proficiency types
		return entities.ProficiencyTypeUnknown

	}
}
func apiClassToClass(input *apiEntities.Class) *entities.Class {
	return &entities.Class{
		Key:                      input.Key,
		Name:                     input.Name,
		HitDie:                   input.HitDie,
		ProficiencyChoices:       apiChoicesToChoices(input.ProficiencyChoices),
		Proficiencies:            apiReferenceItemsToReferenceItems(input.Proficiencies),
		StartingEquipmentChoices: apiChoicesToChoices(input.StartingEquipmentOptions),
		StartingEquipment:        apiStartingEquipmentsToStartingEquipments(input.StartingEquipment),
	}
}

func apiStartingEquipmentToStartingEquipment(input *apiEntities.StartingEquipment) *entities.StartingEquipment {
	return &entities.StartingEquipment{
		Quantity:  input.Quantity,
		Equipment: apiReferenceItemToReferenceItem(input.Equipment),
	}
}

func apiStartingEquipmentsToStartingEquipments(input []*apiEntities.StartingEquipment) []*entities.StartingEquipment {
	output := make([]*entities.StartingEquipment, len(input))
	for i, apiStartingEquipment := range input {
		output[i] = apiStartingEquipmentToStartingEquipment(apiStartingEquipment)
	}

	return output
}
func apiEquipmentInterfaceToEquipment(input dnd5e.EquipmentInterface) entities.Equipment {
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

func apiWeaponToWeapon(input *apiEntities.Weapon) *entities.Weapon {
	return &entities.Weapon{
		Base: entities.BasicEquipment{
			Key:    input.Key,
			Name:   input.Name,
			Weight: input.Weight,
			Cost:   apiCostToCost(input.Cost),
		},
		WeaponCategory:  input.WeaponCategory,
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

func apiArmorToArmor(input *apiEntities.Armor) *entities.Armor {
	return &entities.Armor{
		Base: entities.BasicEquipment{
			Key:    input.Key,
			Name:   input.Name,
			Weight: input.Weight,
			Cost:   apiCostToCost(input.Cost),
		},
		ArmorClass: &entities.ArmorClass{
			Base:     input.ArmorClass.Base,
			DexBonus: input.ArmorClass.DexBonus,
			// MaxBonus is not available from the API, so medium armor limits
			// are handled in the AC calculator fallback logic
		},
		StealthDisadvantage: input.StealthDisadvantage,
	}
}

func apiEquipmentToEquipment(input *apiEntities.Equipment) *entities.BasicEquipment {
	return &entities.BasicEquipment{
		Key:    input.Key,
		Name:   input.Name,
		Weight: input.Weight,
		Cost:   apiCostToCost(input.Cost),
	}
}

func apiCostToCost(input *apiEntities.Cost) *entities.Cost {
	return &entities.Cost{
		Quantity: input.Quantity,
		Unit:     input.Unit,
	}
}

func apiChoicesToChoices(input []*apiEntities.ChoiceOption) []*entities.Choice {
	output := make([]*entities.Choice, len(input))
	for i, apiChoice := range input {
		output[i] = apiChoiceOptionToChoice(apiChoice)
	}

	return output
}

func apiChoiceOptionToChoice(input *apiEntities.ChoiceOption) *entities.Choice {
	if input == nil {
		return nil
	}

	if input.OptionList == nil {
		return nil
	}

	output := make([]entities.Option, len(input.OptionList.Options))

	for i, apiProficiency := range input.OptionList.Options {
		output[i] = apiOptionToOption(apiProficiency)
	}

	// Choice created successfully

	return &entities.Choice{
		Count:   input.ChoiceCount,
		Name:    input.Description,
		Type:    apiChoiceTypeToChoiceType(input.ChoiceType),
		Key:     "choice",
		Options: output,
	}
}

func apiChoiceTypeToChoiceType(input string) entities.ChoiceType {
	switch input {
	case "proficiencies":
		return entities.ChoiceTypeProficiency
	case "equipment":
		return entities.ChoiceTypeEquipment
	case "languages":
		return entities.ChoiceTypeLanguage
	default:
		// Silently handle unknown choice types
		return entities.ChoiceTypeUnset
	}
}

func apiOptionToOption(input apiEntities.Option) entities.Option {
	switch input.GetOptionType() {
	case apiEntities.OptionTypeReference:
		item, ok := input.(*apiEntities.ReferenceOption)
		if !ok || item.Reference == nil {
			return nil
		}

		return &entities.ReferenceOption{
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

		return &entities.CountedReferenceOption{
			Count:     item.Count,
			Reference: apiReferenceItemToReferenceItem(item.Reference),
		}
	case apiEntities.OptionTypeMultiple:
		item, ok := input.(*apiEntities.MultipleOption)
		if !ok || item.Items == nil {
			return nil
		}

		options := make([]entities.Option, len(item.Items))
		for i, apiOption := range item.Items {
			options[i] = apiOptionToOption(apiOption)
		}

		return &entities.MultipleOption{
			Items: options,
		}
	default:
		// Silently handle unknown option types
		return nil
	}
}

func apiReferenceItemsToReferenceItems(input []*apiEntities.ReferenceItem) []*entities.ReferenceItem {
	output := make([]*entities.ReferenceItem, len(input))
	for i, apiReferenceItem := range input {
		output[i] = apiReferenceItemToReferenceItem(apiReferenceItem)
	}

	return output
}
func apiReferenceItemToReferenceItem(input *apiEntities.ReferenceItem) *entities.ReferenceItem {
	return &entities.ReferenceItem{
		Key:  input.Key,
		Name: input.Name,
		Type: typeStringToReferenceType(input.Type),
	}
}

func typeStringToReferenceType(input string) entities.ReferenceType {
	switch input {
	case "equipment":
		return entities.ReferenceTypeEquipment
	case "proficiencies":
		return entities.ReferenceTypeProficiency
	case "languages":
		return entities.ReferenceTypeLanguage
	case "ability-scores":
		return entities.ReferenceTypeAbilityScore
	case "skills":
		return entities.ReferenceTypeSkill
	case "weapon-properties":
		return entities.ReferenceTypeWeaponProperty
	default:
		// Don't log for known but unhandled types
		if input != "2014" {
			log.Println("Unknown reference type: ", input)
		}
		return entities.ReferenceTypeUnset
	}
}
