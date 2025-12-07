package characters

import (
	"fmt"
	"maps"
	"math/rand"
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CharacterRegistry struct {
	db                  *gorm.DB
	teamKeys            mapset.Set[string]
	classKeys           map[string]mapset.Set[string]
	characters          map[string]Character
	secondaryCharacters map[string]Character
	classAssigners      map[string]map[string]ClassAssigner
}

func CharacterRegistryProvider(i *do.Injector) (*CharacterRegistry, error) {
	db, err := do.Invoke[*gorm.DB](i)
	if err != nil {
		return nil, err
	}
	return &CharacterRegistry{
		db:                  db,
		characters:          make(map[string]Character),
		secondaryCharacters: make(map[string]Character),
		classAssigners:      make(map[string]map[string]ClassAssigner),
	}, nil
}

func (ct *CharacterRegistry) registerCharacter(label string, c *Character, primary bool) error {
	if c == nil {
		return errors.New("character cannot be nil")
	}
	var table string
	var list map[string]Character
	if primary {
		table = "characters"
		list = ct.characters
	} else {
		table = "secondary_characters"
		list = ct.secondaryCharacters
	}
	if _, ok := list[label]; ok {
		return errors.New(fmt.Sprintf("Duplicate character registration %s", label))
	}
	dbCharacter := &DbCharacter{Label: label}
	result := ct.db.
		Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "label"}},
			DoNothing: true,
		}).
		Create(&dbCharacter)
	if result.Error != nil {
		return result.Error
	}
	c.DbCharacter = dbCharacter
	list[label] = *c
	return nil
}

func (ct *CharacterRegistry) RegisterPrimary(label string, c *Character) error {
	return ct.registerCharacter(label, c, true)
}

func (ct *CharacterRegistry) GetPrimaries(ids *[]int) []Character {
	if ids == nil {
		characters := make([]Character, 0, len(ct.characters))
		for _, c := range ct.characters {
			characters = append(characters, c)
		}
		return characters
	}

	characters := make([]Character, 0)
	for _, c := range ct.characters {
		if slices.Contains(*ids, c.Id) {
			characters = append(characters, c)
		}
	}
	return characters
}

func (ct *CharacterRegistry) RegisterSecondary(label string, c *Character) error {
	return ct.registerCharacter(label, c, false)
}

func (ct *CharacterRegistry) RegisterClassAssigner(team string, class string, assigner ClassAssigner) error {
	var teamMap map[string]ClassAssigner
	if existingMap, ok := ct.classAssigners[team]; ok {
		teamMap = existingMap
	} else {
		existingMap = make(map[string]ClassAssigner)
		ct.classAssigners[team] = existingMap
	}
	if _, ok := teamMap[class]; ok {
		return errors.New(fmt.Sprintf("Duplicate class assigner registration %s", class))
	}
	teamMap[class] = assigner
	return nil
}

func (ct *CharacterRegistry) Teams() []string {
	return ct.teamKeys.ToSlice()
}

func (ct *CharacterRegistry) AssignedPrimaryCharacter(team, class string, score GameScore) (c Character, extraRequested bool, err error) {
	teamMap, ok := ct.classAssigners[team]
	if !ok {
		err = errors.New(fmt.Sprintf("`%s` team does not exist", team))
		return
	}
	if class != "" {
		if assigner, ok := teamMap[class]; ok {
			return assigner(score)
		} else {
			err = errors.New(fmt.Sprintf("`%s` class in `%s` team does not exist", class, team))
			return
		}
	} else {
		classKeys := slices.Collect(maps.Keys(teamMap))
		return teamMap[classKeys[rand.Intn(len(classKeys))]](score)
	}

}
