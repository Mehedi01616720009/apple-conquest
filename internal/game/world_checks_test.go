package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"
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

func TestStartingCastleDeterminesDistinctRulerColor(t *testing.T) {
	g, err := NewGame("Test Ruler", "green", "Mosul", 42)
	if err != nil {
		t.Fatal(err)
	}
	if g.PlayerColor != "teal" {
		t.Fatalf("player color = %q, want inherited Mosul color teal", g.PlayerColor)
	}
	seen := make(map[string]bool)
	for _, name := range CastleNames() {
		colorName := RulerColorForCastle(name)
		if seen[colorName] {
			t.Fatalf("castle color %q is assigned more than once", colorName)
		}
		seen[colorName] = true
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
	for range 4 {
		g.TickOnce()
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

func TestSpeedPresetsMatchRequestedDayLengths(t *testing.T) {
	g := newTestGame(t)
	for preset, expected := range map[string]int{"slower": 25, "slow": 20, "normal": 15, "fast": 10, "faster": 5} {
		if err := g.SetSpeed(preset); err != nil {
			t.Fatalf("SetSpeed(%q) = %v", preset, err)
		}
		if g.SpeedSeconds != expected {
			t.Fatalf("SpeedSeconds for %q = %d, want %d", preset, g.SpeedSeconds, expected)
		}
		if got := g.TickStepDuration(); got != time.Duration(expected)*time.Second/10 {
			t.Fatalf("TickStepDuration for %q = %v, want %v", preset, got, time.Duration(expected)*time.Second/10)
		}
	}
}

func TestArmyMovementAndAttackQueue(t *testing.T) {
	g := newTestGame(t)
	g.Castles["Cairo"].Troops = 120
	g.Castles["Alexandria"].Owner = g.PlayerName
	g.Castles["Jerusalem"].Owner = "King Baldwin IV"
	g.Relations["King Baldwin IV"] = War
	if err := g.Attack("Cairo", "Jerusalem", 40); err != nil {
		t.Fatal(err)
	}
	if len(g.Orders) != 1 {
		t.Fatalf("queued orders = %d, want 1", len(g.Orders))
	}
	if got := g.Orders[0].Type; got != "attack" {
		t.Fatalf("queued type = %q, want attack", got)
	}
	if got := g.Orders[0].TravelDays; got != 2 {
		t.Fatalf("attack travel days = %d, want 2", got)
	}
	if err := g.MoveArmy("Cairo", "Alexandria", 25); err != nil {
		t.Fatal(err)
	}
	if len(g.Orders) != 2 {
		t.Fatalf("queue length after move = %d, want 2", len(g.Orders))
	}
}

func TestRequestArmyNeedsAdjacentPlayerCastleAndSendsFifteenPercent(t *testing.T) {
	var accepted, rejected bool
	for seed := int64(1); seed <= 100 && (!accepted || !rejected); seed++ {
		g := newTestGame(t)
		g.rng = rand.New(rand.NewSource(seed))
		source := g.Castles["Alexandria"]
		source.Owner = "Ally"
		source.Troops = 100
		g.Relations["Ally"] = Alliance
		if err := g.RequestArmy(source.Name); err != nil {
			t.Fatal(err)
		}
		if len(g.Orders) == 1 {
			accepted = true
			if g.Orders[0].Type != "reinforcement" || g.Orders[0].Troops != 15 || g.Orders[0].To != "Cairo" {
				t.Fatalf("reinforcement order = %+v, want 15 troops to adjacent Cairo", g.Orders[0])
			}
			if source.Troops != 85 {
				t.Fatalf("source troops = %d, want 85 reserved", source.Troops)
			}
			for range 3 {
				g.TickOnce()
			}
			if g.Castles["Cairo"].Troops != 115 {
				t.Fatalf("destination troops = %d, want 115 after shipment", g.Castles["Cairo"].Troops)
			}
		} else {
			rejected = true
			if source.Troops != 100 {
				t.Fatalf("rejected request changed source troops to %d", source.Troops)
			}
		}
	}
	if !accepted || !rejected {
		t.Fatalf("seeded request trials: accepted=%t rejected=%t, want both outcomes", accepted, rejected)
	}

	g := newTestGame(t)
	g.Castles["Cairo"].Owner = "Other ruler"
	g.Castles["Alexandria"].Owner = "Ally"
	g.Relations["Ally"] = Alliance
	if err := g.RequestArmy("Alexandria"); err == nil {
		t.Fatal("request should fail when no neighboring castle belongs directly to the player")
	}
}

func TestVassalCanOnlyDeclareWarByCrossingRebellionThreshold(t *testing.T) {
	g := newTestGame(t)
	vassal := g.Castles["Alexandria"]
	vassal.Owner = "Ally"
	parent := g.Castles["Cairo"]
	parent.Gold, parent.Wood, parent.Food, parent.Troops = 100, 100, 100, 100
	vassal.Gold, vassal.Wood, vassal.Food, vassal.Troops = 80, 70, 70, 60
	g.Vassals["Ally"] = true
	g.VassalParents["Ally"] = g.PlayerName
	g.Relations["Ally"] = Peace
	for range 100 {
		g.aiConsiderWar(vassal)
	}
	if g.Relations["Ally"] == War {
		t.Fatal("vassal independently declared war before meeting rebellion thresholds")
	}
	g.TickOnce()
	if !g.Vassals["Ally"] || g.Relations["Ally"] == War {
		t.Fatal("vassal rebelled while resource or army was not above 70% of the parent")
	}

	vassal.Gold, vassal.Wood, vassal.Food, vassal.Troops = 90, 90, 90, 71
	g.TickOnce()
	if g.Vassals["Ally"] || g.Relations["Ally"] != War || g.Relations[g.PlayerName] != War {
		t.Fatal("vassal did not break away and declare war after exceeding both 70% thresholds")
	}
}

func TestAllianceRequiresComparableStrength(t *testing.T) {
	g := newTestGame(t)
	g.rng = rand.New(rand.NewSource(1))
	cairo := g.Castles["Cairo"]
	alexandria := g.Castles["Alexandria"]
	cairo.Gold, cairo.Wood, cairo.Food = 100, 100, 100
	cairo.Troops = 100
	alexandria.Gold, alexandria.Wood, alexandria.Food = 100, 100, 100
	alexandria.Troops = 20
	if err := g.Diplomacy("ally", "Alexandria"); err == nil && g.Relations[alexandria.Owner] == Alliance {
		t.Fatal("alliance should reject an overmatched proposal")
	}
}

func TestVassalRequiresRelativeAdvantage(t *testing.T) {
	g := newTestGame(t)
	g.rng = rand.New(rand.NewSource(1))
	cairo := g.Castles["Cairo"]
	alexandria := g.Castles["Alexandria"]
	cairo.Gold, cairo.Wood, cairo.Food = 90, 90, 90
	cairo.Troops = 100
	alexandria.Gold, alexandria.Wood, alexandria.Food = 80, 80, 80
	alexandria.Troops = 60
	if err := g.Diplomacy("vassal", "Alexandria"); err == nil && g.Vassals[alexandria.Owner] {
		t.Fatal("vassal should be rejected unless the player exceeds the target by at least the required margins")
	}
}

func TestAIPlayersCanDeclareWarOnEachOther(t *testing.T) {
	g := newTestGame(t)
	cairo := g.Castles["Cairo"]
	alexandria := g.Castles["Alexandria"]
	cairo.Owner = "Sultan Salahuddin Ayyubi"
	alexandria.Owner = "Emir Kazi Fadil"
	g.Relations["Sultan Salahuddin Ayyubi"] = Peace
	g.Relations["Emir Kazi Fadil"] = Peace
	for range 100 {
		g.aiConsiderWar(alexandria)
		if g.Relations["Sultan Salahuddin Ayyubi"] == War || g.Relations["Emir Kazi Fadil"] == War {
			return
		}
	}
	if g.Relations["Emir Kazi Fadil"] != War && g.Relations["Sultan Salahuddin Ayyubi"] != War {
		t.Fatal("AI rulers should be able to declare war on neighboring AI rulers as well as the player")
	}
}
