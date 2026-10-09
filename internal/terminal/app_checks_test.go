package terminal

import (
	"strings"
	"testing"

	"apple-conquest/internal/game"
)

func TestCommandLoopActionsAndMap(t *testing.T) {
	world, err := game.NewGame("Test Ruler", "green", "Cairo", 42)
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	app := &App{Game: world, Out: &output}
	if err := app.Execute("diplomacy war 5"); err != nil {
		t.Fatal(err)
	}
	if err := app.Execute("attack 1 5 100"); err != nil {
		t.Fatal(err)
	}
	if world.Castles["Jerusalem"].Owner != world.PlayerName {
		t.Fatal("numbered attack command did not capture Jerusalem")
	}
	if err := app.Execute("map"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "+------------------------------------+") {
		t.Fatal("map output is missing bordered castle panels")
	}
	if !strings.Contains(output.String(), "Control: PLAYER") || !strings.Contains(output.String(), "Ruler: Test Ruler") {
		t.Fatal("map output is missing player control details")
	}
	mapHasAlignedCastles := false
	for _, line := range strings.Split(output.String(), "\n") {
		if strings.Contains(line, "01  Cairo") && strings.Contains(line, "02  Alexandria") {
			mapHasAlignedCastles = true
			break
		}
	}
	if !mapHasAlignedCastles {
		t.Fatal("map output does not align paired castles on the same row")
	}
	for _, name := range game.CastleNames() {
		if !strings.Contains(output.String(), name) {
			t.Errorf("map output does not contain %s", name)
		}
	}
}

func TestInvalidCommandDoesNotChangeResources(t *testing.T) {
	world, err := game.NewGame("Test Ruler", "green", "Cairo", 42)
	if err != nil {
		t.Fatal(err)
	}
	startingGold := world.Castles["Cairo"].Gold
	app := &App{Game: world, Out: &strings.Builder{}}
	if err := app.Execute("build 1 unknown"); err == nil {
		t.Fatal("expected an invalid building to return an error")
	}
	if world.Castles["Cairo"].Gold != startingGold {
		t.Fatal("invalid command changed the castle's resources")
	}
}

func TestWithdrawCommandMovesGarrisonIntoFieldArmy(t *testing.T) {
	world, err := game.NewGame("Test Ruler", "green", "Cairo", 42)
	if err != nil {
		t.Fatal(err)
	}
	app := &App{Game: world, Out: &strings.Builder{}}
	if err := app.Execute("garrison 1 40"); err != nil {
		t.Fatal(err)
	}
	if err := app.Execute("withdraw 1 15"); err != nil {
		t.Fatal(err)
	}
	cairo := world.Castles["Cairo"]
	if cairo.Troops != 75 || cairo.Garrison != 25 {
		t.Fatalf("field/garrison = %d/%d, want 75/25", cairo.Troops, cairo.Garrison)
	}
}

func TestStatusShowsArmyLevelAndSeparateForces(t *testing.T) {
	world, err := game.NewGame("Test Ruler", "green", "Cairo", 42)
	if err != nil {
		t.Fatal(err)
	}
	cairo := world.Castles["Cairo"]
	cairo.ForgeLevel = 1
	cairo.Troops = 60
	cairo.Garrison = 40
	var output strings.Builder
	app := &App{Game: world, Out: &output}
	app.RenderStatus()
	if !strings.Contains(output.String(), "Army level 2: 60 field, 40 garrison") {
		t.Fatal("status does not show army level and separate field/garrison counts")
	}
}
