package terminal

import (
	"fmt"
	"strings"

	"apple-conquest/internal/game"
)

func (a *App) RenderMap() {
	fmt.Fprintln(a.Out, "\nREGIONAL MAP ------------------------------------")
	rows := [][]string{
		{"", "08 Edessa", "09 Aleppo", "10 Mosul"},
		{"06 Tripoli", "", "04 Damascus", ""},
		{"", "07 Acre", "05 Jerusalem", "03 Kerak"},
		{"", "", "01 Cairo", "02 Alexandria"},
	}
	for _, row := range rows {
		fmt.Fprint(a.Out, "  ")
		for _, label := range row {
			if label == "" {
				fmt.Fprint(a.Out, strings.Repeat(" ", 19))
				continue
			}
			index := label[:2]
			name := strings.TrimSpace(label[2:])
			castle := a.Game.Castles[name]
			ownerMark := "A"
			if castle.Owner == a.Game.PlayerName {
				ownerMark = "P"
			} else if a.Game.Vassals[castle.Owner] {
				ownerMark = "V"
			}
			text := fmt.Sprintf("[%s %-10s %s]", index, name, ownerMark)
			if a.Color {
				if ownerMark == "YOU" {
					text = colorCodes[a.Game.PlayerColor] + text + "\033[0m"
				} else if ownerMark == "VAS" {
					text = "\033[36m" + text + "\033[0m"
				} else {
					text = "\033[31m" + text + "\033[0m"
				}
			}
			fmt.Fprintf(a.Out, "%-19s", text)
		}
		fmt.Fprintln(a.Out)
	}
	fmt.Fprintln(a.Out, "  P = player   V = vassal-held   A = independent ruler")
	fmt.Fprintln(a.Out, "  Neighboring castles:")
	for _, name := range a.Game.Order {
		castle := a.Game.Castles[name]
		fmt.Fprintf(a.Out, "    %02d %-12s -> %s\n", indexOf(name)+1, name, strings.Join(castle.Neighbors, ", "))
	}
}

func (a *App) RenderStatus() {
	fmt.Fprintln(a.Out, "\nREALM STATUS ------------------------------------")
	if a.Color {
		fmt.Fprintf(a.Out, "Ruler: %s%s\033[0m | Day: %d | Time: %s\n", colorCodes[a.Game.PlayerColor], a.Game.PlayerName, a.Game.Tick/10, pauseLabel(a.Game.Paused))
	} else {
		fmt.Fprintf(a.Out, "Ruler: %s | Day: %d | Time: %s\n", a.Game.PlayerName, a.Game.Tick/10, pauseLabel(a.Game.Paused))
	}
	fmt.Fprintln(a.Out, "-------------------------------------------------")
	for _, name := range a.Game.Order {
		c := a.Game.Castles[name]
		if c.Owner != a.Game.PlayerName {
			continue
		}
		fmt.Fprintf(a.Out, "%02d %-12s | Gold %4d Wood %4d Food %4d | Pop %4d\n", indexOf(name)+1, c.Name, c.Gold, c.Wood, c.Food, c.Population)
		fmt.Fprintf(a.Out, "   Army %d (%d garrison) Artillery %d Forge %d\n", c.Troops, c.Garrison, c.Artillery, c.ForgeLevel)
		fmt.Fprintf(a.Out, "   Buildings: %s\n", buildingSummary(c))
	}
	fmt.Fprintf(a.Out, "Vassals: %d | Outcome: %s\n", len(a.Game.Vassals), a.Game.Outcome)
}

func (a *App) RenderCastles() {
	fmt.Fprintln(a.Out, "\nCASTLES -----------------------------------------")
	for _, name := range a.Game.Order {
		c := a.Game.Castles[name]
		status := "Independent"
		if c.Owner == a.Game.PlayerName {
			status = "Your realm"
		} else if a.Game.Vassals[c.Owner] {
			status = "Vassal"
		}
		relation := a.Game.Relations[c.Owner]
		if relation == "" {
			relation = game.Peace
		}
		fmt.Fprintf(a.Out, "%02d %-12s | %-30s | %-11s | relation: %s\n", indexOf(name)+1, c.Name, c.Owner, status, relation)
	}
}

func (a *App) RenderEvents() {
	fmt.Fprintln(a.Out, "\nRECENT EVENTS -----------------------------------")
	start := len(a.Game.Events) - 10
	if start < 0 {
		start = 0
	}
	for _, event := range a.Game.Events[start:] {
		fmt.Fprintf(a.Out, "  * %s\n", event)
	}
}

func indexOf(name string) int {
	for index, candidate := range game.CastleNames() {
		if candidate == name {
			return index
		}
	}
	return 0
}

func buildingSummary(c *game.Castle) string {
	buildings := []game.Building{game.Farm, game.House, game.Sawmill, game.Market, game.Barracks, game.Artillery, game.Forge}
	parts := make([]string, 0, len(buildings))
	for _, building := range buildings {
		if count := c.Buildings[building]; count > 0 {
			parts = append(parts, fmt.Sprintf("%s x%d", building, count))
		}
	}
	return strings.Join(parts, ", ")
}

func pauseLabel(paused bool) string {
	if paused {
		return "paused"
	}
	return "running"
}
