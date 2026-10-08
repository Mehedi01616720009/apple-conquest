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