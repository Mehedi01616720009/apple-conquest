package terminal

import (
	"bytes"
	"strings"
	"testing"

	"apple-conquest/internal/game"
)

func TestRunBuildsBuildingAndContinuesAfterInvalidCommand(t *testing.T) {
	input := strings.NewReader("build saw mill\ntrain archer\nstatus\nquit\n")
	var output bytes.Buffer

	if err := Run(input, &output, game.NewGame()); err != nil {
		t.Fatal(err)
	}
	result := output.String()
	if !strings.Contains(result, "saw mill 1") {
		t.Fatalf("output does not show built saw mill:\n%s", result)
	}
	if !strings.Contains(result, "Error: build a barracks before training units") {
		t.Fatalf("output does not show training error:\n%s", result)
	}
	if !strings.Contains(result, "Goodbye.") {
		t.Fatalf("terminal did not process quit after an invalid command:\n%s", result)
	}
}
