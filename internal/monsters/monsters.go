package monsters

import (
	"math/rand"
)

// Monster represents a universal monster with facts
type Monster struct {
	Name        string
	Description string
	Facts       []string
	Origin      string
	FirstApp    string
}

var monsters = []Monster{
	{
		Name:        "Dracula",
		Description: "The legendary vampire count from Transylvania",
		Origin:      "Romanian folklore",
		FirstApp:    "Bram Stoker's Dracula (1897)",
		Facts: []string{
			"Dracula can transform into a bat, wolf, or mist",
			"He is repelled by garlic, crosses, and holy water",
			"Dracula must rest in his coffin with Transylvanian soil",
			"He casts no reflection and has no shadow",
			"The character was inspired by Vlad the Impaler",
			"Dracula has superhuman strength and can control weather",
		},
	},
	{
		Name:        "Frankenstein's Monster",
		Description: "The tragic creature created by Dr. Victor Frankenstein",
		Origin:      "Literary fiction",
		FirstApp:    "Frankenstein by Mary Shelley (1818)",
		Facts: []string{
			"The monster is never given a name in the original novel",
			"He was created from assembled body parts",
			"The monster is highly intelligent and teaches himself to read",
			"He seeks love and acceptance but is rejected by society",
			"Mary Shelley wrote Frankenstein when she was only 18 years old",
			"The monster's first request was for a female companion",
		},
	},
	{
		Name:        "The Wolf Man",
		Description: "A man cursed to transform into a werewolf",
		Origin:      "European folklore",
		FirstApp:    "The Wolf Man (1941 film)",
		Facts: []string{
			"Werewolves transform during a full moon",
			"Silver bullets are the most effective weapon against werewolves",
			"The curse can be passed through a bite or scratch",
			"Wolfsbane is a plant that can repel werewolves",
			"Larry Talbot was the original Wolf Man character",
			"The transformation is often depicted as painful",
		},
	},
	{
		Name:        "The Mummy",
		Description: "An ancient Egyptian priest brought back to life",
		Origin:      "Ancient Egyptian mythology",
		FirstApp:    "The Mummy (1932 film)",
		Facts: []string{
			"Imhotep was the name of the original Universal Mummy",
			"The Mummy seeks to resurrect his lost love",
			"Ancient curses were believed to protect tombs",
			"The discovery of King Tut's tomb inspired the Mummy films",
			"Mummies were preserved through a 70-day process",
			"The Mummy can command supernatural powers",
		},
	},
	{
		Name:        "The Creature from the Black Lagoon",
		Description: "An amphibious humanoid from the Amazon",
		Origin:      "Film original",
		FirstApp:    "Creature from the Black Lagoon (1954)",
		Facts: []string{
			"The Creature is also known as Gill-man",
			"He is a prehistoric evolutionary link",
			"The Creature can breathe both underwater and on land",
			"He becomes fascinated by humans, particularly women",
			"The film was originally shot in 3D",
			"The Creature represents nature's mystery and danger",
		},
	},
	{
		Name:        "The Invisible Man",
		Description: "A scientist who discovers the secret of invisibility",
		Origin:      "Literary fiction",
		FirstApp:    "The Invisible Man by H.G. Wells (1897)",
		Facts: []string{
			"Dr. Griffin is driven mad by his condition",
			"Invisibility cannot be reversed in the original story",
			"The Invisible Man must be completely nude to be fully invisible",
			"Fog, smoke, or snow can reveal his presence",
			"The character explores themes of power and corruption",
			"He wears bandages and dark glasses to appear visible",
		},
	},
	{
		Name:        "The Phantom of the Opera",
		Description: "A disfigured musical genius haunting the Paris Opera House",
		Origin:      "Literary fiction",
		FirstApp:    "The Phantom of the Opera by Gaston Leroux (1909)",
		Facts: []string{
			"The Phantom's real name is Erik",
			"He lives in the catacombs beneath the opera house",
			"Erik is a master of music, magic, and architecture",
			"His face is severely disfigured",
			"He falls in love with Christine Daaé, a young soprano",
			"The story is partly based on real events at the Paris Opera",
		},
	},
	{
		Name:        "The Hunchback of Notre Dame",
		Description: "The deformed bell-ringer of Notre Dame Cathedral",
		Origin:      "Literary fiction",
		FirstApp:    "Notre-Dame de Paris by Victor Hugo (1831)",
		Facts: []string{
			"His name is Quasimodo, meaning 'half-formed'",
			"He is deaf from years of ringing the cathedral bells",
			"Quasimodo has a kind heart despite his appearance",
			"He is deeply loyal to Esmeralda",
			"The novel helped save Notre Dame from demolition",
			"Quasimodo has superhuman strength and agility",
		},
	},
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
		if monster.Name == name {
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
