package dnd5e

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"

	apiEntities "github.com/fadedpez/dnd5e-api/entities"

	internal "github.com/KirkDiggler/dnd-bot-discord/internal"
	"github.com/fadedpez/dnd5e-api/clients/dnd5e"
)

// TODO: add context to functions
type client struct {
	client dnd5e.Interface
}

type Config struct {
	HttpClient *http.Client
}

func New(cfg *Config) (Client, error) {
	if cfg == nil {
		return nil, internal.NewMissingParamError("cfg")
	}

	dndClient, err := dnd5e.NewDND5eAPI(&dnd5e.DND5eAPIConfig{
		Client: cfg.HttpClient,
	})
	if err != nil {
		return nil, err
	}

	return &client{
		client: dndClient,
	}, nil
}

func (c *client) ListClasses() ([]*rulebook.Class, error) {
	response, err := c.client.ListClasses()
	if err != nil {
		return nil, err
	}

	return apiReferenceItemsToClasses(response), nil
}

func (c *client) ListRaces() ([]*rulebook.Race, error) {
	response, err := c.client.ListRaces()
	if err != nil {
		return nil, err
	}

	return apiReferenceItemsToRaces(response), nil
}

func (c *client) GetRace(key string) (*rulebook.Race, error) {
	response, err := c.client.GetRace(key)
	if err != nil {
		return nil, err
	}

	race := apiRaceToRace(response)

	return race, nil
}

func (c *client) GetClass(key string) (*rulebook.Class, error) {
	response, err := c.client.GetClass(key)
	if err != nil {
		return nil, err
	}

	return apiClassToClass(response), nil
}

func (c *client) GetProficiency(key string) (*rulebook.Proficiency, error) {
	if key == "" {
		return nil, internal.NewMissingParamError("GetProficiency.key")
	}

	response, err := c.doGetProficiency(key)
	if err != nil {
		return nil, err
	}

	return apiProficiencyToProficiency(response), nil
}

func (c *client) GetEquipment(key string) (equipment.Equipment, error) {
	if key == "" {
		return nil, internal.NewMissingParamError("GetEquipment.key")
	}

	response, err := c.client.GetEquipment(key)
	if err != nil {
		return nil, err
	}

	return apiEquipmentInterfaceToEquipment(response), nil
}

func (c *client) doGetProficiency(key string) (*apiEntities.Proficiency, error) {
	response, err := c.client.GetProficiency(key)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *client) GetMonster(key string) (*combat.MonsterTemplate, error) {
	monsterTemplate, err := c.client.GetMonster(key)
	if err != nil {
		return nil, err
	}

	return apiToMonsterTemplate(monsterTemplate), nil
}

func (c *client) GetEquipmentByCategory(category string) ([]equipment.Equipment, error) {
	// Get equipment category data
	categoryData, err := c.client.GetEquipmentCategory(category)
	if err != nil {
		return nil, err
	}

	// Fetch each piece of equipment
	equipmentSlice := make([]equipment.Equipment, 0, len(categoryData.Equipment))
	for _, ref := range categoryData.Equipment {
		if ref.Key != "" {
			equip, err := c.GetEquipment(ref.Key)
			if err != nil {
				// Log error but continue with other equipment
				continue
			}
			equipmentSlice = append(equipmentSlice, equip)
		}
	}

	return equipmentSlice, nil
}

// GetClassFeatures returns features for a class at a specific level
func (c *client) GetClassFeatures(classKey string, level int) ([]*rulebook.CharacterFeature, error) {
	// Get the class level data which includes features
	classLevel, err := c.client.GetClassLevel(classKey, level)
	if err != nil {
		return nil, err
	}

	// Convert feature references to CharacterFeatures
	features := make([]*rulebook.CharacterFeature, 0, len(classLevel.Features))
	for _, featureRef := range classLevel.Features {
		if featureRef.Key != "" {
			// For now, we just create basic features from the reference
			// In the future, we might want to fetch full feature details
			feature := &rulebook.CharacterFeature{
				Key:         featureRef.Key,
				Name:        featureRef.Name,
				Description: "", // Would need to fetch full feature for description
				Type:        rulebook.FeatureTypeClass,
				Level:       level,
				Source:      classKey,
			}
			features = append(features, feature)
		}
	}

	return features, nil
}

// ListMonstersByCR returns monsters within a challenge rating range
func (c *client) ListMonstersByCR(minCR, maxCR float32) ([]*combat.MonsterTemplate, error) {
	// The API only supports filtering by exact CR, not range
	// So we need to fetch monsters for each CR value in the range
	crValues := getCRValuesInRange(minCR, maxCR)

	monsters := make([]*combat.MonsterTemplate, 0)
	processedKeys := make(map[string]bool) // Track processed monsters to avoid duplicates

	for _, cr := range crValues {
		crFloat64 := float64(cr)
		input := &dnd5e.ListMonstersInput{
			ChallengeRating: &crFloat64,
		}

		// Get monster references for this CR
		monsterRefs, err := c.client.ListMonstersWithFilter(input)
		if err != nil {
			log.Printf("Failed to list monsters for CR %f: %v", cr, err)
			continue
		}

		// Fetch each monster's details
		for _, ref := range monsterRefs {
			if ref.Key != "" && !processedKeys[ref.Key] {
				monster, err := c.client.GetMonster(ref.Key)
				if err != nil {
					log.Printf("Failed to get monster %s: %v", ref.Key, err)
					continue
				}
				if monster != nil {
					template := apiToMonsterTemplate(monster)
					if template != nil {
						monsters = append(monsters, template)
						processedKeys[ref.Key] = true
					}
				}
			}
		}
	}

	return monsters, nil
}

// getCRValuesInRange returns all standard CR values within the given range
func getCRValuesInRange(minCR, maxCR float32) []float32 {
	// Standard D&D 5e CR values
	allCRs := []float32{0, 0.125, 0.25, 0.5, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30}

	var result []float32
	for _, cr := range allCRs {
		if cr >= minCR && cr <= maxCR {
			result = append(result, cr)
		}
	}
	return result
}

// ListEquipment returns all equipment
func (c *client) ListEquipment() ([]equipment.Equipment, error) {
	// Get list of all equipment references
	equipmentRefs, err := c.client.ListEquipment()
	if err != nil {
		return nil, err
	}

	// Fetch each piece of equipment
	equipmentValue := make([]equipment.Equipment, 0, len(equipmentRefs))
	for _, ref := range equipmentRefs {
		if ref.Key != "" {
			equip, err := c.GetEquipment(ref.Key)
			if err != nil {
				// Log error but continue with other equipment
				log.Printf("Failed to get equipment %s: %v", ref.Key, err)
				continue
			}
			if equip != nil {
				equipmentValue = append(equipmentValue, equip)
			}
		}
	}

	return equipmentValue, nil
}

func apiToMonsterTemplate(input *apiEntities.Monster) *combat.MonsterTemplate {
	if input == nil {
		return nil
	}

	return &combat.MonsterTemplate{
		Key:             input.Key,
		Name:            input.Name,
		Type:            input.Type,
		ArmorClass:      input.ArmorClass,
		HitPoints:       input.HitPoints,
		HitDice:         input.HitDice,
		ChallengeRating: input.ChallengeRating,
		Actions:         apisToMonsterActions(input.MonsterActions),
	}
}

func apiToDamage(input *apiEntities.Damage) *damage.Damage {
	a := strings.Split(input.DamageDice, "+")
	dice := input.DamageDice
	var bonus, diceValue, diceCount int
	var err error
	if len(a) == 2 {
		bonus, err = strconv.Atoi(a[1])
		if err != nil {
			log.Printf("Unknown dice format %s", input.DamageDice)
			return nil
		}
		dice = a[0]
	}

	b := strings.Split(dice, "d")
	if len(b) == 2 {
		diceCount, err = strconv.Atoi(b[0])
		if err != nil {
			log.Printf("Unknown dice format %s", input.DamageDice)
			return nil
		}
		diceValue, err = strconv.Atoi(b[1])
		if err != nil {
			log.Printf("Unknown dice format %s", input.DamageDice)
			return nil
		}
	}

	// TODO: add damage type
	return &damage.Damage{
		DiceCount: diceCount,
		DiceSize:  diceValue,
		Bonus:     bonus,
	}
}

func apisToDamages(input []*apiEntities.Damage) []*damage.Damage {
	if input == nil {
		return nil
	}

	var damages []*damage.Damage
	for _, d := range input {
		damages = append(damages, apiToDamage(d))
	}

	return damages
}

func apisToMonsterActions(input []*apiEntities.MonsterAction) []*combat.MonsterAction {
	if input == nil {
		return nil
	}

	var monsterActions []*combat.MonsterAction
	for _, ma := range input {
		monsterActions = append(monsterActions, apiToMonsterAction(ma))
	}

	return monsterActions
}

func apiToMonsterAction(input *apiEntities.MonsterAction) *combat.MonsterAction {
	if input == nil {
		return nil
	}

	return &combat.MonsterAction{
		Name:        input.Name,
		Description: input.Description,
		AttackBonus: input.AttackBonus,
		Damage:      apisToDamages(input.Damage),
	}
}
