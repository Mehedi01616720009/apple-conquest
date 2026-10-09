package terminal

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"apple-conquest/internal/game"
)

func (a *App) RenderMap() {
	fmt.Fprintln(a.Out, "\nREGIONAL MAP ------------------------------------")
	rows := [][2]string{
		{"Edessa", "Aleppo"},
		{"Mosul", "Damascus"},
		{"Tripoli", "Acre"},
		{"Jerusalem", "Kerak"},
		{"Cairo", "Alexandria"},
	}
	for _, row := range rows {
		left := a.mapCard(row[0])
		right := a.mapCard(row[1])
		border := "+" + strings.Repeat("-", 36) + "+"
		fmt.Fprintf(a.Out, "  %s  %s\n", border, border)
		for line := range left {
			fmt.Fprintf(a.Out, "  %s  %s\n", a.renderMapCellLine(left[line]), a.renderMapCellLine(right[line]))
		}
		fmt.Fprintf(a.Out, "  %s  %s\n", border, border)
		fmt.Fprintln(a.Out)
	}
	fmt.Fprintln(a.Out, "REALM COLORS -------------------------------------")
	for index, name := range a.Game.Order {
		owner := a.Game.Castles[name].OriginalRuler
		colorName := game.RulerColor(owner)
		label := owner
		if colorName == a.Game.PlayerColor {
			label = "You: " + a.Game.PlayerName
		}
		if index%5 == 0 {
			fmt.Fprint(a.Out, "  ")
		}
		if a.Color {
			fmt.Fprintf(a.Out, "%s%-22s\033[0m", colorCodes[colorName], truncateMapText(label, 20))
		} else {
			fmt.Fprintf(a.Out, "%-22s", truncateMapText(label, 20))
		}
		if index%5 == 4 || index == len(a.Game.Order)-1 {
			fmt.Fprintln(a.Out)
		} else {
			fmt.Fprint(a.Out, "  ")
		}
	}
	fmt.Fprintln(a.Out, "NEIGHBORS ----------------------------------------")
	for _, name := range a.Game.Order {
		castle := a.Game.Castles[name]
		fmt.Fprintf(a.Out, "%02d %-12s -> %s\n", indexOf(name)+1, name, strings.Join(castle.Neighbors, ", "))
	}
}

func (a *App) mapCard(name string) []mapCellLine {
	castle := a.Game.Castles[name]
	control := "AI"
	ownerColor := game.RulerColor(castle.Owner)
	if castle.Owner == a.Game.PlayerName {
		control = "PLAYER"
		ownerColor = a.Game.PlayerColor
	} else if a.Game.Vassals[castle.Owner] {
		control = "VASSAL"
	}
	relation := a.Game.Relations[castle.Owner]
	if relation == "" {
		relation = game.Peace
	}
	return []mapCellLine{
		{text: fmt.Sprintf("%02d  %s", indexOf(name)+1, castle.Name)},
		{},
		{text: "Ruler: " + castle.Owner, colorValue: castle.Owner, color: colorCodes[ownerColor]},
		{text: "Control: " + control},
		{text: "Relation: " + string(relation)},
		{},
	}
}

type mapCellLine struct {
	text       string
	colorValue string
	color      string
}

func (a *App) renderMapCellLine(line mapCellLine) string {
	text := truncateMapText(line.text, 34)
	display := text
	if a.Color && line.colorValue != "" {
		if prefix, ok := strings.CutPrefix(text, "Ruler: "); ok {
			display = fmt.Sprintf("Ruler: %s%s\033[0m", line.color, prefix)
		}
	}
	padding := 34 - utf8.RuneCountInString(text)
	if padding < 0 {
		padding = 0
	}
	return "| " + display + strings.Repeat(" ", padding) + " |"
}

func truncateMapText(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-3]) + "..."
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
		fmt.Fprintf(a.Out, "   Army level %d: %d field, %d garrison | Artillery %d\n", c.ForgeLevel+1, c.Troops, c.Garrison, c.Artillery)
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
