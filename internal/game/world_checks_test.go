package game

import "testing"

func newTestGame(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame("Test Ruler", "green", "Cairo", 42)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestNewGameAssignsTenCastles(t *testing.T) {
	g := newTestGame(t)
	if got := len(g.Castles); got != 10 {
		t.Fatalf("got %d castles, want 10", got)
	}
	if got := g.Castles["Cairo"].Owner; got != g.PlayerName {
		t.Fatalf("starting castle owner = %q, want player", got)
	}
	if got := g.Castles["Alexandria"].Owner; got != "Emir Kazi Fadil" {
		t.Fatalf("AI castle owner = %q, want Emir Kazi Fadil", got)
	}
}

func TestTickAdvancesEconomyDeterministically(t *testing.T) {
	first := newTestGame(t)
	second := newTestGame(t)
	for range 10 {
		first.TickOnce()
		second.TickOnce()
	}
	if first.Tick != 10 {
		t.Fatalf("tick = %d, want 10", first.Tick)
	}
	for _, name := range first.Order {
		a, b := first.Castles[name], second.Castles[name]
		if a.Gold != b.Gold || a.Wood != b.Wood || a.Food != b.Food || a.Troops != b.Troops {
			t.Fatalf("simulation diverged for %s with the same seed", name)
		}
	}
}

func TestAITroopsGrowOverTime(t *testing.T) {
	g := newTestGame(t)
	startingTroops := g.Castles["Alexandria"].Troops
	for range 5 {
		g.TickOnce()
	}
	if got := g.Castles["Alexandria"].Troops; got <= startingTroops {
		t.Fatalf("AI troops = %d after five ticks, want more than %d", got, startingTroops)
	}
}

func TestBuildTrainAndGarrison(t *testing.T) {
	g := newTestGame(t)
	cairo := g.Castles["Cairo"]
	startingGold := cairo.Gold
	if err := g.Build("Cairo", Farm); err != nil {
		t.Fatal(err)
	}
	if cairo.Buildings[Farm] != 2 || cairo.Gold >= startingGold {
		t.Fatal("building did not increase farm count and spend gold")
	}
	if err := g.Train("Cairo", "soldiers", 10); err != nil {
		t.Fatal(err)
	}
	if cairo.Troops != 110 {
		t.Fatalf("troops = %d, want 110", cairo.Troops)
	}
	if err := g.SetGarrison("Cairo", 25); err != nil {
		t.Fatal(err)
	}
	if cairo.Garrison != 25 || cairo.Troops != 85 {
		t.Fatalf("army/garrison = %d/%d, want 85/25", cairo.Troops, cairo.Garrison)
	}
	if err := g.Build("Alexandria", Farm); err == nil {
		t.Fatal("expected management of AI castle to fail")
	}
}

func TestWarAttackAndVassalVictory(t *testing.T) {
	g := newTestGame(t)
	if err := g.Diplomacy("war", "Jerusalem"); err != nil {
		t.Fatal(err)
	}
	if err := g.Attack("Cairo", "Jerusalem", 100); err != nil {
		t.Fatal(err)
	}
	if got := g.Castles["Jerusalem"].Owner; got != g.PlayerName {
		t.Fatalf("Jerusalem owner = %q, want player after capture", got)
	}
	for _, c := range g.Castles {
		if c.Owner != g.PlayerName {
			g.Vassals[c.Owner] = true
		}
	}
	if got := g.CheckOutcome(); got != Victory {
		t.Fatalf("outcome = %q, want victory", got)
	}
}

func TestRepelledAttackReturnsSurvivors(t *testing.T) {
	g := newTestGame(t)
	attacker := g.Castles["Cairo"]
	defender := g.Castles["Jerusalem"]
	attacker.Troops = 20
	defender.Troops = 200
	g.resolveBattle(attacker, defender, 10)
	if attacker.Troops != 15 {
		t.Fatalf("attacker survivors = %d, want 15", attacker.Troops)
	}
}

func TestLossWhenPlayerHasNoCastle(t *testing.T) {
	g := newTestGame(t)
	for _, c := range g.Castles {
		if c.Owner == g.PlayerName {
			c.Owner = "Sultan Salahuddin Ayyubi"
		}
	}
	if got := g.CheckOutcome(); got != Defeat {
		t.Fatalf("outcome = %q, want defeat", got)
	}
}
