package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type Building string

const (
	Farm      Building = "farm"
	House     Building = "house"
	Sawmill   Building = "sawmill"
	Market    Building = "market"
	Barracks  Building = "barracks"
	Artillery Building = "artillery"
	Forge     Building = "forge"
)

type Relation string

const (
	Peace    Relation = "peace"
	Truce    Relation = "truce"
	Rivalry  Relation = "rivalry"
	War      Relation = "war"
	Alliance Relation = "alliance"
)

type Outcome string

const (
	Ongoing Outcome = "ongoing"
	Victory Outcome = "victory"
	Defeat  Outcome = "defeat"
)

type Castle struct {
	Name          string
	OriginalRuler string
	Owner         string
	Neighbors     []string
	Population    int
	Gold          int
	Wood          int
	Food          int
	Troops        int
	Garrison      int
	Artillery     int
	ForgeLevel    int
	Buildings     map[Building]int
}

type Game struct {
	PlayerName  string
	PlayerColor string
	Castles     map[string]*Castle
	Order       []string
	Relations   map[string]Relation
	Vassals     map[string]bool
	Events      []string
	Tick        int
	Paused      bool
	Outcome     Outcome
	rng         *rand.Rand
	aiActivity  []string
}

type castleSeed struct {
	name, ruler              string
	neighbors                []string
	population               int
	gold, wood, food, troops int
}

var castleSeeds = []castleSeed{
	{"Cairo", "Sultan Salahuddin Ayyubi", []string{"Alexandria", "Damascus", "Jerusalem"}, 520, 150, 90, 130, 100},
	{"Alexandria", "Emir Kazi Fadil", []string{"Cairo"}, 430, 120, 90, 120, 75},
	{"Kerak", "Lord Reynald De Chatillon", []string{"Jerusalem", "Damascus"}, 360, 105, 75, 95, 70},
	{"Damascus", "Sultan Nur Al Din Zengi", []string{"Cairo", "Jerusalem", "Kerak", "Acre", "Aleppo"}, 560, 155, 110, 145, 110},
	{"Jerusalem", "King Baldwin IV", []string{"Cairo", "Damascus", "Kerak", "Acre"}, 500, 140, 100, 125, 95},
	{"Tripoli", "Lord Reymond III", []string{"Acre", "Aleppo"}, 390, 110, 90, 105, 80},
	{"Acre", "Baron Balian of Ibelin", []string{"Tripoli", "Damascus", "Jerusalem"}, 450, 130, 95, 120, 90},
	{"Edessa", "Joscelin III", []string{"Aleppo", "Mosul"}, 330, 95, 85, 100, 65},
	{"Aleppo", "Emir Muzaffar Uddin Gokbori", []string{"Edessa", "Mosul", "Damascus", "Tripoli"}, 470, 135, 105, 125, 95},
	{"Mosul", "Emir Saifuddin Zengi", []string{"Edessa", "Aleppo"}, 490, 145, 110, 135, 100},
}

var buildingCosts = map[Building]struct{ gold, wood int }{
	Farm:      {35, 12},
	House:     {45, 18},
	Sawmill:   {40, 10},
	Market:    {55, 15},
	Barracks:  {70, 30},
	Artillery: {100, 45},
	Forge:     {85, 35},
}

func NewGame(playerName, playerColor, startingCastle string, seed int64) (*Game, error) {
	playerName = strings.TrimSpace(playerName)
	if playerName == "" {
		return nil, fmt.Errorf("ruler name cannot be empty")
	}

	start, err := ResolveCastle(startingCastle)
	if err != nil {
		return nil, err
	}

	g := &Game{
		PlayerName:  playerName,
		PlayerColor: playerColor,
		Castles:     make(map[string]*Castle, len(castleSeeds)),
		Relations:   make(map[string]Relation),
		Vassals:     make(map[string]bool),
		Outcome:     Ongoing,
		rng:         rand.New(rand.NewSource(seed)),
	}
	for _, seed := range castleSeeds {
		c := &Castle{
			Name:          seed.name,
			OriginalRuler: seed.ruler,
			Owner:         seed.ruler,
			Neighbors:     append([]string(nil), seed.neighbors...),
			Population:    seed.population,
			Gold:          seed.gold,
			Wood:          seed.wood,
			Food:          seed.food,
			Troops:        seed.troops,
			Buildings:     map[Building]int{Farm: 1, House: 1, Sawmill: 1, Market: 1, Barracks: 1},
		}
		g.Castles[c.Name] = c
		g.Order = append(g.Order, c.Name)
		g.Relations[c.Owner] = Peace
	}
	g.Castles[start].Owner = playerName
	g.addEvent(fmt.Sprintf("%s has taken command of %s.", playerName, start))
	return g, nil
}

func ResolveCastle(value string) (string, error) {
	value = strings.TrimSpace(value)
	if index, err := strconv.Atoi(value); err == nil && index >= 1 && index <= len(castleSeeds) {
		return castleSeeds[index-1].name, nil
	}
	for _, seed := range castleSeeds {
		if strings.EqualFold(value, seed.name) {
			return seed.name, nil
		}
	}
	return "", fmt.Errorf("unknown castle %q; use a castle number or name", value)
}

func CastleNames() []string {
	names := make([]string, 0, len(castleSeeds))
	for _, seed := range castleSeeds {
		names = append(names, seed.name)
	}
	return names
}

func (g *Game) Castle(name string) (*Castle, error) {
	resolved, err := ResolveCastle(name)
	if err != nil {
		return nil, err
	}
	return g.Castles[resolved], nil
}

func (g *Game) PlayerControls(c *Castle) bool {
	return c.Owner == g.PlayerName || g.Vassals[c.Owner]
}

func (g *Game) CheckOutcome() Outcome {
	directCastles := 0
	controlledCastles := 0
	for _, c := range g.Castles {
		if c.Owner == g.PlayerName {
			directCastles++
		}
		if g.PlayerControls(c) {
			controlledCastles++
		}
	}
	if directCastles == 0 {
		g.Outcome = Defeat
	} else if controlledCastles == len(g.Castles) {
		g.Outcome = Victory
	}
	return g.Outcome
}

func (g *Game) addEvent(message string) {
	g.Events = append(g.Events, message)
}
