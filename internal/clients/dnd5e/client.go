package dnd5e

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
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

func (c *client) ListClasses() ([]*entities.Class, error) {
	response, err := c.client.ListClasses()
	if err != nil {
		return nil, err
	}

	return apiReferenceItemsToClasses(response), nil
}

func (c *client) ListRaces() ([]*entities.Race, error) {
	response, err := c.client.ListRaces()
	if err != nil {
		return nil, err
	}

	return apiReferenceItemsToRaces(response), nil
}

func (c *client) GetRace(key string) (*entities.Race, error) {
	response, err := c.client.GetRace(key)
	if err != nil {
		return nil, err
	}

	race := apiRaceToRace(response)

	return race, nil
}

func (c *client) GetClass(key string) (*entities.Class, error) {
	response, err := c.client.GetClass(key)
	if err != nil {
		return nil, err
	}

	return apiClassToClass(response), nil
}

func (c *client) GetProficiency(key string) (*entities.Proficiency, error) {
	if key == "" {
		return nil, internal.NewMissingParamError("GetProficiency.key")
	}

	response, err := c.doGetProficiency(key)
	if err != nil {
		return nil, err
	}

	return apiProficiencyToProficiency(response), nil
}

func (c *client) GetEquipment(key string) (entities.Equipment, error) {
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

func (c *client) GetMonster(key string) (*entities.MonsterTemplate, error) {
	monsterTemplate, err := c.client.GetMonster(key)
	if err != nil {
		return nil, err
	}

	return apiToMonsterTemplate(monsterTemplate), nil
}

func apiToMonsterTemplate(input *apiEntities.Monster) *entities.MonsterTemplate {
	if input == nil {
		return nil
	}

	return &entities.MonsterTemplate{
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
	var dice string = input.DamageDice
	var bonus, diceValue, diceCount int
	if len(a) == 2 {
		bonus, _ = strconv.Atoi(a[1])
		dice = a[0]
	}

	b := strings.Split(dice, "d")
	if len(b) == 2 {
		diceCount, _ = strconv.Atoi(b[0])
		diceValue, _ = strconv.Atoi(b[1])
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

func apisToMonsterActions(input []*apiEntities.MonsterAction) []*entities.MonsterAction {
	if input == nil {
		return nil
	}

	var monsterActions []*entities.MonsterAction
	for _, ma := range input {
		monsterActions = append(monsterActions, apiToMonsterAction(ma))
	}

	return monsterActions
}

func apiToMonsterAction(input *apiEntities.MonsterAction) *entities.MonsterAction {
	if input == nil {
		return nil
	}

	return &entities.MonsterAction{
		Name:        input.Name,
		Description: input.Description,
		AttackBonus: input.AttackBonus,
		Damage:      apisToDamages(input.Damage),
	}
}
