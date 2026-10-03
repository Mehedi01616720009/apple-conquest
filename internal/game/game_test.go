package game

import "testing"

func TestNewGameHasStarterSettlement(t *testing.T) {
	state := NewGame().State()
	if state.Population != 3 || state.Buildings[Castle] != 1 || state.Buildings[Hut] != 1 {
		t.Fatalf("unexpected starter state: %+v", state)
	}
	if state.Resources[Wood] != 30 || state.Resources[Food] != 30 || state.Resources[Gold] != 15 {
		t.Fatalf("unexpected starter resources: %+v", state.Resources)
	}
}

func TestBuildChargesConfiguredCost(t *testing.T) {
	game := NewGame()
	if err := game.Build(WheatField); err != nil {
		t.Fatal(err)
	}
	state := game.State()
	if state.Resources[Wood] != 18 || state.Buildings[WheatField] != 1 {
		t.Fatalf("build did not apply its cost and building count: %+v", state)
	}
}

func TestRejectedBuildDoesNotMutateState(t *testing.T) {
	game := NewGame()
	game.state.Resources[Gold] = 5
	before := game.State()
	if err := game.Build(Barracks); err == nil {
		t.Fatal("Build(Barracks) succeeded without enough gold")
	}
	assertSameState(t, before, game.State())
}

func TestAdvanceTurnProducesBeforeFeeding(t *testing.T) {
	game := NewGame()
	game.state.Buildings[WheatField] = 1
	game.state.Buildings[SawMill] = 1
	game.state.Buildings[TaxCollector] = 1
	if err := game.AdvanceTurn(); err != nil {
		t.Fatal(err)
	}
	state := game.State()
	if state.Resources[Food] != 32 || state.Resources[Wood] != 34 || state.Resources[Gold] != 18 {
		t.Fatalf("unexpected resources after turn: %+v", state.Resources)
	}
	if state.Population != 4 {
		t.Fatalf("population = %d, want 4", state.Population)
	}
}

func TestStarvationRemovesOneCivilianPerMissingFood(t *testing.T) {
	game := NewGame()
	game.state.Population = 4
	game.state.Resources[Food] = 2
	if err := game.AdvanceTurn(); err != nil {
		t.Fatal(err)
	}
	state := game.State()
	if state.Population != 2 || state.Resources[Food] != 0 {
		t.Fatalf("unexpected starvation result: population=%d food=%d", state.Population, state.Resources[Food])
	}
}

func TestStarvationCanEndGame(t *testing.T) {
	game := NewGame()
	game.state.Population = 1
	game.state.Resources[Food] = 0
	if err := game.AdvanceTurn(); err != nil {
		t.Fatal(err)
	}
	if got := game.State().Status; got != Lost {
		t.Fatalf("status = %s, want lost", got)
	}
}

func TestTradeRequiresMarketAndUsesFixedRate(t *testing.T) {
	game := NewGame()
	if err := game.Trade(Wood, Gold, 2); err == nil {
		t.Fatal("Trade succeeded without a market")
	}
	game.state.Buildings[Market] = 1
	if err := game.Trade(Wood, Gold, 4); err != nil {
		t.Fatal(err)
	}
	state := game.State()
	if state.Resources[Wood] != 26 || state.Resources[Gold] != 17 {
		t.Fatalf("unexpected trade result: %+v", state.Resources)
	}
}

func TestTrainingRequiresBarracks(t *testing.T) {
	game := NewGame()
	if err := game.Train(Archer); err == nil {
		t.Fatal("Train succeeded without a barracks")
	}
	if got := game.State().Units[Archer]; got != 0 {
		t.Fatalf("archer count = %d after rejected training, want 0", got)
	}
}

func TestTrainingFinalUnitCompletesObjective(t *testing.T) {
	game := NewGame()
	game.state.Population = populationToWin
	game.state.Buildings[Barracks] = 1
	game.state.Units[Archer] = 1
	game.state.Units[Infantry] = 1
	game.state.Resources[Food] = 5
	game.state.Resources[Gold] = 6
	if err := game.Train(Cavalry); err != nil {
		t.Fatal(err)
	}
	if got := game.State().Status; got != Won {
		t.Fatalf("status = %s, want won", got)
	}
}

func TestStarterScenarioCanReachVictory(t *testing.T) {
	game := NewGame()
	if err := game.Build(WheatField); err != nil {
		t.Fatal(err)
	}
	if err := game.Build(SawMill); err != nil {
		t.Fatal(err)
	}
	advanceTurns(t, game, 2)
	if err := game.Build(TaxCollector); err != nil {
		t.Fatal(err)
	}
	advanceTurns(t, game, 3)
	if err := game.Build(Hut); err != nil {
		t.Fatal(err)
	}
	advanceTurns(t, game, 4)
	if err := game.Build(Barracks); err != nil {
		t.Fatal(err)
	}
	for _, unit := range []Unit{Archer, Infantry, Cavalry} {
		if err := game.Train(unit); err != nil {
			t.Fatal(err)
		}
	}
	if got := game.State().Status; got != Won {
		t.Fatalf("status = %s, want won", got)
	}
}

func TestStateReturnsIndependentMaps(t *testing.T) {
	game := NewGame()
	state := game.State()
	state.Resources[Wood] = 0
	state.Buildings[Hut] = 0
	if actual := game.State(); actual.Resources[Wood] != 30 || actual.Buildings[Hut] != 1 {
		t.Fatalf("external state mutation changed game: %+v", actual)
	}
}

func advanceTurns(t *testing.T, game *Game, count int) {
	t.Helper()
	for range count {
		if err := game.AdvanceTurn(); err != nil {
			t.Fatal(err)
		}
	}
}

func assertSameState(t *testing.T, want, got State) {
	t.Helper()
	if want.Turn != got.Turn || want.Population != got.Population || want.Status != got.Status {
		t.Fatalf("state changed: before=%+v after=%+v", want, got)
	}
	for resource, amount := range want.Resources {
		if got.Resources[resource] != amount {
			t.Fatalf("resource %s changed: before=%d after=%d", resource, amount, got.Resources[resource])
		}
	}
	for building, count := range want.Buildings {
		if got.Buildings[building] != count {
			t.Fatalf("building %s changed: before=%d after=%d", building, count, got.Buildings[building])
		}
	}
}
