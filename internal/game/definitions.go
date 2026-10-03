package game

type cost struct {
	wood int
	food int
	gold int
}

var buildingCosts = map[Building]cost{
	Castle:       {wood: 30},
	Hut:          {wood: 10},
	SawMill:      {wood: 12},
	WheatField:   {wood: 12},
	TaxCollector: {wood: 12},
	Market:       {wood: 15},
	Barracks:     {wood: 20, gold: 10},
}

var trainingCosts = map[Unit]cost{
	Archer:   {food: 2, gold: 2},
	Infantry: {food: 3, gold: 3},
	Cavalry:  {food: 5, gold: 6},
}

const (
	startingWood        = 30
	startingFood        = 30
	startingGold        = 15
	startingPopulation  = 3
	housingPerHut       = 5
	foodPerCivilian     = 1
	woodPerSawMill      = 4
	foodPerWheatField   = 5
	goldPerTaxCollector = 3
	populationToWin     = 8
)

type tradeRate struct {
	input  int
	output int
}

var marketRates = map[[2]Resource]tradeRate{
	{Wood, Gold}: {input: 2, output: 1},
	{Food, Gold}: {input: 3, output: 1},
	{Gold, Wood}: {input: 1, output: 2},
	{Gold, Food}: {input: 1, output: 3},
}

func NewGame() *Game {
	return &Game{state: State{
		Resources: map[Resource]int{
			Wood: startingWood,
			Food: startingFood,
			Gold: startingGold,
		},
		Buildings: map[Building]int{
			Castle: 1,
			Hut:    1,
		},
		Units:      make(map[Unit]int),
		Population: startingPopulation,
		Status:     Ongoing,
	}}
}
