package game

import "fmt"

func (game *Game) Build(building Building) error {
	if err := game.requireOngoing(); err != nil {
		return err
	}
	buildingCost, ok := buildingCosts[building]
	if !ok {
		return fmt.Errorf("unknown building %d", building)
	}
	if building == Castle && game.state.Buildings[Castle] > 0 {
		return fmt.Errorf("a castle has already been built")
	}
	if err := game.canAfford(buildingCost); err != nil {
		return err
	}
	game.pay(buildingCost)
	game.state.Buildings[building]++
	return nil
}

func (game *Game) Train(unit Unit) error {
	if err := game.requireOngoing(); err != nil {
		return err
	}
	unitCost, ok := trainingCosts[unit]
	if !ok {
		return fmt.Errorf("unknown unit %d", unit)
	}
	if game.state.Buildings[Barracks] == 0 {
		return fmt.Errorf("build a barracks before training units")
	}
	if err := game.canAfford(unitCost); err != nil {
		return err
	}
	game.pay(unitCost)
	game.state.Units[unit]++
	game.updateStatus()
	return nil
}

func (game *Game) Trade(from, to Resource, amount int) error {
	if err := game.requireOngoing(); err != nil {
		return err
	}
	if game.state.Buildings[Market] == 0 {
		return fmt.Errorf("build a market before trading")
	}
	if amount <= 0 {
		return fmt.Errorf("trade amount must be positive")
	}
	rate, ok := marketRates[[2]Resource{from, to}]
	if !ok {
		return fmt.Errorf("trading %s for %s is not available", from, to)
	}
	if amount%rate.input != 0 {
		return fmt.Errorf("trade amount must be a multiple of %d %s", rate.input, from)
	}
	if game.state.Resources[from] < amount {
		return fmt.Errorf("not enough %s", from)
	}
	game.state.Resources[from] -= amount
	game.state.Resources[to] += amount / rate.input * rate.output
	return nil
}

func (game *Game) AdvanceTurn() error {
	if err := game.requireOngoing(); err != nil {
		return err
	}
	game.state.Turn++
	game.state.Resources[Wood] += game.state.Buildings[SawMill] * woodPerSawMill
	game.state.Resources[Food] += game.state.Buildings[WheatField] * foodPerWheatField
	game.state.Resources[Gold] += game.state.Buildings[TaxCollector] * goldPerTaxCollector

	foodNeeded := game.state.Population * foodPerCivilian
	fed := game.state.Resources[Food] >= foodNeeded
	if fed {
		game.state.Resources[Food] -= foodNeeded
	} else {
		missingFood := foodNeeded - game.state.Resources[Food]
		game.state.Resources[Food] = 0
		game.state.Population -= missingFood
		if game.state.Population < 0 {
			game.state.Population = 0
		}
	}

	if fed && game.state.Population < game.state.HousingCapacity() {
		game.state.Population++
	}
	game.updateStatus()
	return nil
}

func (game *Game) canAfford(required cost) error {
	if game.state.Resources[Wood] < required.wood ||
		game.state.Resources[Food] < required.food ||
		game.state.Resources[Gold] < required.gold {
		return fmt.Errorf("not enough resources")
	}
	return nil
}

func (game *Game) pay(required cost) {
	game.state.Resources[Wood] -= required.wood
	game.state.Resources[Food] -= required.food
	game.state.Resources[Gold] -= required.gold
}

func (game *Game) requireOngoing() error {
	if game.state.Status != Ongoing {
		return fmt.Errorf("game is already %s", game.state.Status)
	}
	return nil
}

func (game *Game) updateStatus() {
	if game.state.Population == 0 {
		game.state.Status = Lost
		return
	}
	if game.state.Population >= populationToWin &&
		game.state.Units[Archer] > 0 &&
		game.state.Units[Infantry] > 0 &&
		game.state.Units[Cavalry] > 0 {
		game.state.Status = Won
	}
}
