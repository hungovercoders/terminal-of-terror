package monsters

import (
	_ "embed"
	"encoding/json"
	"log"
	"math/rand"
	"strings"
)

// Monster represents a universal monster with facts
type Monster struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Facts       []string `json:"facts"`
	Origin      string   `json:"origin"`
	FirstApp    string   `json:"firstApp"`
	ASCII       string   `json:"ascii"`
}

//go:embed monsters.json
var monstersJSON []byte

var monsters []Monster

func init() {
	if err := json.Unmarshal(monstersJSON, &monsters); err != nil {
		log.Fatalf("Failed to load monsters data: %v", err)
	}
}

// GetAllMonsters returns all available monsters
func GetAllMonsters() []Monster {
	return monsters
}

// GetRandomMonster returns a random monster
func GetRandomMonster() Monster {
	return monsters[rand.Intn(len(monsters))]
}

// GetMonsterByName returns a monster by name (case-insensitive)
func GetMonsterByName(name string) *Monster {
	for _, monster := range monsters {
		if strings.EqualFold(monster.Name, name) {
			return &monster
		}
	}
	return nil
}

// GetRandomFact returns a random fact from a random monster
func GetRandomFact() (string, string) {
	monster := GetRandomMonster()
	fact := monster.Facts[rand.Intn(len(monster.Facts))]
	return monster.Name, fact
}
