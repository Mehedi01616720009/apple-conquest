package game

import "fmt"

type Resource uint8

const (
	Wood Resource = iota
	Food
	Gold
)

func (resource Resource) String() string {
	switch resource {
	case Wood:
		return "wood"
	case Food:
		return "food"
	case Gold:
		return "gold"
	default:
		return fmt.Sprintf("resource(%d)", resource)
	}
}

type Building uint8

const (
	Castle Building = iota
	Hut
	SawMill
	WheatField
	TaxCollector
	Market
	Barracks
)

func (building Building) String() string {
	switch building {
	case Castle:
		return "castle"
	case Hut:
		return "hut"
	case SawMill:
		return "saw mill"
	case WheatField:
		return "wheat field"
	case TaxCollector:
		return "tax collector"
	case Market:
		return "market"
	case Barracks:
		return "barracks"
	default:
		return fmt.Sprintf("building(%d)", building)
	}
}

type Unit uint8

const (
	Archer Unit = iota
	Infantry
	Cavalry
)

func (unit Unit) String() string {
	switch unit {
	case Archer:
		return "archer"
	case Infantry:
		return "infantry"
	case Cavalry:
		return "cavalry"
	default:
		return fmt.Sprintf("unit(%d)", unit)
	}
}

type Status uint8

const (
	Ongoing Status = iota
	Won
	Lost
)

func (status Status) String() string {
	switch status {
	case Ongoing:
		return "ongoing"
	case Won:
		return "won"
	case Lost:
		return "lost"
	default:
		return fmt.Sprintf("status(%d)", status)
	}
}

type State struct {
	Turn       int
	Resources  map[Resource]int
	Buildings  map[Building]int
	Population int
	Units      map[Unit]int
	Status     Status
}

func (state State) HousingCapacity() int {
	return state.Buildings[Hut] * housingPerHut
}

func (state State) clone() State {
	copy := state
	copy.Resources = make(map[Resource]int, len(state.Resources))
	for resource, amount := range state.Resources {
		copy.Resources[resource] = amount
	}
	copy.Buildings = make(map[Building]int, len(state.Buildings))
	for building, count := range state.Buildings {
		copy.Buildings[building] = count
	}
	copy.Units = make(map[Unit]int, len(state.Units))
	for unit, count := range state.Units {
		copy.Units[unit] = count
	}
	return copy
}

type Game struct {
	state State
}

func (game *Game) State() State {
	return game.state.clone()
}
