package desktop

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"apple-conquest/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (a *App) drawSetup(screen *ebiten.Image) {
	screen.Fill(paper)
	vector.DrawFilledRect(screen, 0, 0, canvasWidth, 86, ink, false)
	a.drawText(screen, "APPLE CONQUEST", "", 46, 22, 32, panel)
	a.drawText(screen, "REAL-TIME STRATEGY  /  THE LEVANT", "", 49, 58, 12, color.RGBA{R: 191, G: 202, B: 185, A: 255})
	a.drawText(screen, "A ruler's beginning", "", 53, 132, 26, ink)
	a.drawText(screen, "Choose a name and stronghold. Your banner follows its ruler.", "", 53, 168, 15, mutedInk)

	drawPanel(screen, image.Rect(44, 204, 501, 829))
	a.drawText(screen, "YOUR RULER", "", 78, 233, 12, mutedInk)
	nameBox := image.Rect(78, 270, 466, 329)
	drawInput(a, screen, nameBox, a.name, a.nameFocused)
	a.drawText(screen, "YOUR BANNER", "", 78, 358, 12, mutedInk)
	colorBounds := image.Rect(78, 389, 120, 431)
	vector.DrawFilledRect(screen, float32(colorBounds.Min.X), float32(colorBounds.Min.Y), float32(colorBounds.Dx()), float32(colorBounds.Dy()), colors[game.RulerColorForCastle(a.startingCastle)], true)
	a.drawText(screen, "Inherited from your starting castle", "", 137, 402, 13, mutedInk)
	a.drawText(screen, "STARTING STRONGHOLD", "", 78, 466, 12, mutedInk)
	vector.DrawFilledRect(screen, 78, 497, 388, 112, land, true)
	a.drawText(screen, a.startingCastle, "", 101, 520, 25, ink)
	a.drawText(screen, "Its ruler becomes your first rival.", "", 101, 558, 14, mutedInk)
	vector.DrawFilledRect(screen, 78, 645, 388, 1, lineColor, false)
	a.drawText(screen, "The remaining nine castles begin under AI rulers.", "", 78, 675, 13, mutedInk)
	a.drawText(screen, "The realm advances in real time.", "", 78, 700, 13, mutedInk)
	if a.notice != "" {
		a.drawText(screen, a.notice, "", 78, 752, 13, red)
	}

	drawPanel(screen, image.Rect(529, 204, 1316, 711))
	a.drawText(screen, "CHOOSE YOUR SEAT", "", 566, 232, 12, mutedInk)
	for index, name := range game.CastleNames() {
		column, row := index%2, index/2
		bounds := image.Rect(566+column*353, 267+row*81, 900+column*353, 323+row*81)
		active := name == a.startingCastle
		fill := paper
		if active {
			fill = land
		}
		drawPanelFill(screen, bounds, fill)
		if active {
			vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), 2, green, true)
		}
		a.drawText(screen, fmt.Sprintf("%02d", index+1), "", float64(bounds.Min.X+15), float64(bounds.Min.Y+18), 14, green)
		a.drawText(screen, name, "", float64(bounds.Min.X+52), float64(bounds.Min.Y+17), 17, ink)
	}
	button := image.Rect(948, 747, 1288, 809)
	a.drawButton(screen, button, "ENTER THE CAMPAIGN", green, panel, false)
	a.drawText(screen, "Click any castle, then enter.", "", 985, 821, 12, mutedInk)
}

func (a *App) drawGame(screen *ebiten.Image) {
	screen.Fill(paper)
	vector.DrawFilledRect(screen, 0, 0, canvasWidth, 76, ink, false)
	a.drawText(screen, "APPLE CONQUEST", "", 30, 23, 22, panel)
	a.drawText(screen, fmt.Sprintf("DAY %02d", a.model.Tick/10), "", 280, 27, 14, color.RGBA{R: 224, G: 196, B: 128, A: 255})
	goldTotal, woodTotal, foodTotal := resourceTotals(a.model)
	a.drawText(screen, fmt.Sprintf("GOLD  %d", goldTotal), "", 396, 27, 14, panel)
	a.drawText(screen, fmt.Sprintf("WOOD  %d", woodTotal), "", 540, 27, 14, panel)
	a.drawText(screen, fmt.Sprintf("FOOD  %d", foodTotal), "", 684, 27, 14, panel)
	buttonLabel := "PAUSE"
	if a.model.Paused {
		buttonLabel = "RESUME"
	}
	for index, preset := range []struct {
		label   string
		seconds int
	}{{"SLOWER", 25}, {"SLOW", 20}, {"NORMAL", 15}, {"FAST", 10}, {"FASTER", 5}} {
		fill := color.RGBA{R: 239, G: 233, B: 214, A: 255}
		textColor := ink
		if a.model.SpeedSeconds == preset.seconds {
			fill, textColor = green, panel
		}
		a.drawButton(screen, image.Rect(820+index*60, 17, 874+index*60, 57), preset.label, fill, textColor, false)
	}
	a.drawButton(screen, image.Rect(1205, 17, 1328, 57), buttonLabel, color.RGBA{R: 53, G: 75, B: 62, A: 255}, panel, false)

	a.drawMap(screen)
	a.drawCastlePanel(screen)
}

func (a *App) drawMap(screen *ebiten.Image) {
	mapBounds := image.Rect(28, 98, 900, 870)
	drawPanelFill(screen, mapBounds, sea)
	a.drawText(screen, "THE LEVANT", "", 57, 120, 13, mutedInk)
	a.drawText(screen, "10 CASTLES  /  CLICK TO INSPECT", "", 695, 120, 11, mutedInk)
	drawRegion(screen)

	for _, name := range a.model.Order {
		castle := a.model.Castles[name]
		start := castlePoints[name]
		for _, neighborName := range castle.Neighbors {
			if castleIndex(name) >= castleIndex(neighborName) {
				continue
			}
			end := castlePoints[neighborName]
			vector.StrokeLine(screen, float32(start.X), float32(start.Y), float32(end.X), float32(end.Y), 2, color.RGBA{R: 164, G: 163, B: 136, A: 255}, true)
		}
	}

	for _, name := range a.model.Order {
		castle := a.model.Castles[name]
		point := castlePoints[name]
		ownerColor := rulerColor(castle.Owner, a.model)
		if castle.Owner == a.model.PlayerName {
			ownerColor = colors[a.model.PlayerColor]
		}
		vector.DrawFilledCircle(screen, float32(point.X), float32(point.Y), 27, ownerColor, true)
		if name == a.selectedCastle {
			vector.StrokeCircle(screen, float32(point.X), float32(point.Y), 33, 3, ink, true)
		}
		vector.StrokeCircle(screen, float32(point.X), float32(point.Y), 27, 2, panel, true)
		army := castle.Troops + castle.Garrison
		a.drawCentered(screen, fmt.Sprintf("%02d", castleIndex(name)+1), float64(point.X), float64(point.Y-9), 17, panel)
		a.drawCentered(screen, name, float64(point.X), float64(point.Y+35), 15, ink)
		a.drawCentered(screen, fmt.Sprintf("L%d  /  %d troops", castle.ForgeLevel+1, army), float64(point.X), float64(point.Y+56), 11, mutedInk)
	}

	for index, castleName := range a.model.Order {
		castle := a.model.Castles[castleName]
		owner := castle.OriginalRuler
		column, row := index%5, index/5
		x, y := 48+column*168, 812+row*25
		ownerColor := colors[game.RulerColor(owner)]
		vector.DrawFilledCircle(screen, float32(x), float32(y+4), 5, ownerColor, true)
		label := owner
		if ownerColor == colors[a.model.PlayerColor] {
			label = "You: " + a.model.PlayerName
		}
		a.drawText(screen, truncate(label, 19), "", float64(x+10), float64(y), 9, ink)
	}
}

func (a *App) drawCastlePanel(screen *ebiten.Image) {
	panelBounds := image.Rect(920, 98, 1332, 870)
	drawPanel(screen, panelBounds)
	castle := a.model.Castles[a.selectedCastle]
	if castle == nil {
		return
	}
	control, controlColor := "INDEPENDENT", rulerColor(castle.Owner, a.model)
	if castle.Owner == a.model.PlayerName {
		control, controlColor = "YOUR CASTLE", colors[a.model.PlayerColor]
	} else if a.model.Vassals[castle.Owner] {
		control = "VASSAL"
	}
	a.drawText(screen, fmt.Sprintf("%02d  %s", castleIndex(castle.Name)+1, castle.Name), "", 950, 124, 23, ink)
	a.drawText(screen, castle.Owner, "", 950, 160, 14, mutedInk)
	a.drawText(screen, control, "", 950, 188, 11, controlColor)
	relation := a.model.Relations[castle.Owner]
	relationLabel := strings.ToUpper(string(relation))
	if castle.Owner == a.model.PlayerName {
		relationLabel = "YOUR REALM"
	} else if relationLabel == "" {
		relationLabel = "PEACE"
	}
	a.drawText(screen, fmt.Sprintf("RELATION  %s", relationLabel), "", 1090, 188, 11, mutedInk)
	vector.DrawFilledRect(screen, 948, 218, 356, 1, lineColor, false)
	a.drawText(screen, fmt.Sprintf("GOLD  %d", castle.Gold), "", 950, 240, 14, gold)
	a.drawText(screen, fmt.Sprintf("WOOD  %d", castle.Wood), "", 1070, 240, 14, color.RGBA{R: 124, G: 88, B: 57, A: 255})
	a.drawText(screen, fmt.Sprintf("FOOD  %d", castle.Food), "", 1191, 240, 14, green)
	a.drawText(screen, fmt.Sprintf("Population  %d", castle.Population), "", 950, 272, 13, mutedInk)
	a.drawText(screen, fmt.Sprintf("Army level %d", castle.ForgeLevel+1), "", 950, 307, 16, ink)
	a.drawText(screen, fmt.Sprintf("%d field  /  %d garrison", castle.Troops, castle.Garrison), "", 950, 335, 13, mutedInk)
	a.drawText(screen, fmt.Sprintf("Artillery  %d", castle.Artillery), "", 950, 361, 13, mutedInk)
	vector.DrawFilledRect(screen, 948, 393, 356, 1, lineColor, false)
	if castle.Owner == a.model.PlayerName {
		a.drawButton(screen, actionRect(0, 0, true), attackOriginLabel(a.attackOrigin, castle.Name), paper, ink, a.attackOrigin == castle.Name)
		a.drawButton(screen, actionRect(1, 0, false), "RECRUIT 10", land, ink, false)
		a.drawButton(screen, actionRect(1, 1, false), "RECRUIT ART", land, ink, false)
		a.drawButton(screen, actionRect(1, 2, false), "BUILD FARM", land, ink, false)
		a.drawButton(screen, actionRect(2, 0, false), "BUILD SAWMILL", land, ink, false)
		a.drawButton(screen, actionRect(2, 1, false), "BUILD MARKET", land, ink, false)
		a.drawButton(screen, actionRect(2, 2, false), "BUILD BARRACKS", land, ink, false)
		a.drawButton(screen, actionRect(3, 0, false), "BUILD ARTILLERY", land, ink, false)
		a.drawButton(screen, actionRect(3, 1, false), "BUILD FORGE", land, ink, false)
		a.drawButton(screen, actionRect(3, 2, false), "UPGRADE ARMY", land, ink, false)
		a.drawButton(screen, actionRect(4, 0, false), "GARRISON +10", land, ink, false)
		a.drawButton(screen, actionRect(4, 1, false), "WITHDRAW 10", land, ink, false)
		a.drawButton(screen, actionRect(4, 2, false), "MOVE ARMY", land, ink, false)
	} else {
		a.drawButton(screen, actionRect(0, 0, false), "DECLARE WAR", color.RGBA{R: 244, G: 224, B: 214, A: 255}, red, false)
		a.drawButton(screen, actionRect(0, 1, false), "TRUCE", land, ink, false)
		a.drawButton(screen, actionRect(1, 0, false), "ALLIANCE", land, ink, false)
		a.drawButton(screen, actionRect(1, 1, false), "VASSAL OFFER", land, ink, false)
		attackEnabled := a.attackOrigin != "" && containsCastle(castle.Neighbors, a.attackOrigin) && a.model.Relations[castle.Owner] == game.War
		attackFill, attackText := land, mutedInk
		if attackEnabled {
			attackFill, attackText = green, panel
		}
		attackLabel := "SELECT ATTACK ORIGIN"
		if a.attackOrigin != "" {
			attackLabel = "ATTACK FROM " + strings.ToUpper(a.attackOrigin)
		}
		a.drawButton(screen, actionRect(2, 0, true), attackLabel, attackFill, attackText, false)
		a.drawButton(screen, actionRect(3, 0, false), "REQUEST ARMY", land, ink, false)
		a.drawButton(screen, actionRect(3, 1, false), "TRADE", land, ink, false)
	}
	if a.notice != "" {
		a.drawText(screen, truncate(a.notice, 56), "", 950, 640, 11, mutedInk)
	}
	vector.DrawFilledRect(screen, 948, 665, 356, 1, lineColor, false)
	a.drawText(screen, "LATEST EVENTS", "", 950, 679, 11, mutedInk)
	events := a.model.Events
	start := max(0, len(events)-5)
	for index, event := range events[start:] {
		a.drawText(screen, truncate(event, 47), "", 950, float64(703+index*25), 11, ink)
	}
	if a.model.Outcome != game.Ongoing {
		outcome := "CAMPAIGN LOST"
		if a.model.Outcome == game.Victory {
			outcome = "REALM UNITED  /  VICTORY"
		}
		vector.DrawFilledRect(screen, 948, 827, 356, 28, ink, true)
		a.drawCentered(screen, outcome, 1126, 832, 12, panel)
	}
}

func drawRegion(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 84, 183, 690, 585, land, true)
	vector.StrokeLine(screen, 90, 240, 280, 200, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
	vector.StrokeLine(screen, 280, 200, 430, 214, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
	vector.StrokeLine(screen, 430, 214, 621, 175, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
	vector.StrokeLine(screen, 621, 175, 780, 207, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
	vector.StrokeLine(screen, 780, 207, 756, 680, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
	vector.StrokeLine(screen, 756, 680, 568, 770, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
	vector.StrokeLine(screen, 568, 770, 320, 752, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
	vector.StrokeLine(screen, 320, 752, 132, 633, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
	vector.StrokeLine(screen, 132, 633, 90, 240, 2, color.RGBA{R: 198, G: 185, B: 151, A: 255}, true)
}

func drawPanel(screen *ebiten.Image, bounds image.Rectangle) {
	drawPanelFill(screen, bounds, panel)
	vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), 1, lineColor, true)
}

func drawPanelFill(screen *ebiten.Image, bounds image.Rectangle, fill color.Color) {
	vector.DrawFilledRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), fill, true)
}

func drawInput(a *App, screen *ebiten.Image, bounds image.Rectangle, value string, focused bool) {
	fill := panel
	border := lineColor
	if focused {
		border = green
	}
	drawPanelFill(screen, bounds, fill)
	vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), 2, border, true)
	a.drawText(screen, value, "left", float64(bounds.Min.X+14), float64(bounds.Min.Y+18), 18, ink)
}

func castleIndex(name string) int {
	for index, castleName := range game.CastleNames() {
		if castleName == name {
			return index
		}
	}
	return -1
}

func attackOriginLabel(origin, current string) string {
	if origin == "" || origin == current {
		return "SET AS ATTACK ORIGIN"
	}
	return "ATTACK ORIGIN  /  " + strings.ToUpper(origin)
}

func rulerColor(owner string, world *game.Game) color.RGBA {
	if owner == world.PlayerName {
		return colors[world.PlayerColor]
	}
	return colors[game.RulerColor(owner)]
}

func resourceTotals(world *game.Game) (goldTotal, woodTotal, foodTotal int) {
	for _, castle := range world.Castles {
		if castle.Owner != world.PlayerName {
			continue
		}
		goldTotal += castle.Gold
		woodTotal += castle.Wood
		foodTotal += castle.Food
	}
	return
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-3]) + "..."
}
