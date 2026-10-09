package game

import (
	"strings"
	"testing"
)

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

func TestAIActivityIsReported(t *testing.T) {
	g := newTestGame(t)
	for range 10 {
		g.TickOnce()
	}
	if !strings.Contains(strings.Join(g.Events, " "), "AI activity") {
		t.Fatal("AI development was not reported in the event log")
	}
}

func TestAdjacentAIEventuallyDeclaresWar(t *testing.T) {
	g := newTestGame(t)
	alexandria := g.Castles["Alexandria"]
	for range 100 {
		g.aiConsiderWar(alexandria)
		if g.Relations[alexandria.Owner] == War {
			return
		}
	}
	t.Fatal("adjacent AI ruler never declared war during repeated decisions")
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

func TestBattleSameLevelCasualties(t *testing.T) {
	g := newTestGame(t)
	attacker := g.Castles["Cairo"]
	defender := g.Castles["Jerusalem"]
	attacker.Troops = 100
	defender.Troops = 120
	g.resolveBattle(attacker, defender, 100)
	if attacker.Troops != 0 {
		t.Fatalf("attacking survivors = %d, want 0", attacker.Troops)
	}
	if defender.Troops != 20 {
		t.Fatalf("defending survivors = %d, want 20", defender.Troops)
	}
}

func TestHigherLevelArmyWinsWithScaledLosses(t *testing.T) {
	g := newTestGame(t)
	attacker := g.Castles["Cairo"]
	defender := g.Castles["Jerusalem"]
	attacker.Troops = 100
	attacker.ForgeLevel = 1
	defender.Troops = 120
	g.resolveBattle(attacker, defender, 100)
	if defender.Owner != attacker.Owner {
		t.Fatalf("castle owner = %q, want attacker %q", defender.Owner, attacker.Owner)
	}
	if defender.Troops != 40 {
		t.Fatalf("attacking survivors in captured castle = %d, want 40", defender.Troops)
	}
	if attacker.Troops != 0 {
		t.Fatalf("source field army = %d, want committed army to leave", attacker.Troops)
	}
}

func TestGarrisonDefendsSeparatelyFromFieldArmy(t *testing.T) {
	g := newTestGame(t)
	attacker := g.Castles["Cairo"]
	defender := g.Castles["Jerusalem"]
	attacker.Troops = 100
	attacker.Garrison = 40
	defender.Troops = 80
	defender.Garrison = 40
	g.resolveBattle(attacker, defender, 100)
	if attacker.Garrison != 40 {
		t.Fatalf("attacker source garrison = %d, want 40", attacker.Garrison)
	}
	if defender.Troops != 20 || defender.Garrison != 0 {
		t.Fatalf("defender field/garrison = %d/%d, want 20/0", defender.Troops, defender.Garrison)
	}
}

func TestRepelledAttackReturnsSurvivors(t *testing.T) {
	g := newTestGame(t)
	attacker := g.Castles["Cairo"]
	defender := g.Castles["Jerusalem"]
	attacker.Troops = 20
	defender.Troops = 200
	g.resolveBattle(attacker, defender, 10)
	if attacker.Troops != 10 {
		t.Fatalf("attacker survivors = %d, want 10", attacker.Troops)
	}
}

func TestWithdrawGarrisonMovesTroopsToFieldArmy(t *testing.T) {
	g := newTestGame(t)
	if err := g.SetGarrison("Cairo", 40); err != nil {
		t.Fatal(err)
	}
	if err := g.WithdrawGarrison("Cairo", 15); err != nil {
		t.Fatal(err)
	}
	cairo := g.Castles["Cairo"]
	if cairo.Troops != 75 || cairo.Garrison != 25 {
		t.Fatalf("field/garrison = %d/%d, want 75/25", cairo.Troops, cairo.Garrison)
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
