package desktop

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"apple-conquest/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	canvasWidth  = 1360
	canvasHeight = 900
)

var (
	ink       = color.RGBA{R: 37, G: 52, B: 45, A: 255}
	mutedInk  = color.RGBA{R: 104, G: 116, B: 103, A: 255}
	paper     = color.RGBA{R: 244, G: 239, B: 226, A: 255}
	panel     = color.RGBA{R: 255, G: 252, B: 244, A: 255}
	lineColor = color.RGBA{R: 205, G: 199, B: 181, A: 255}
	sea       = color.RGBA{R: 222, G: 235, B: 226, A: 255}
	land      = color.RGBA{R: 232, G: 222, B: 196, A: 255}
	green     = color.RGBA{R: 61, G: 112, B: 78, A: 255}
	red       = color.RGBA{R: 188, G: 78, B: 62, A: 255}
	blue      = color.RGBA{R: 64, G: 111, B: 151, A: 255}
	gold      = color.RGBA{R: 193, G: 137, B: 43, A: 255}
	colors    = map[string]color.RGBA{
		"red":     {R: 190, G: 68, B: 56, A: 255},
		"orange":  {R: 213, G: 119, B: 50, A: 255},
		"green":   {R: 62, G: 122, B: 78, A: 255},
		"yellow":  {R: 193, G: 137, B: 43, A: 255},
		"blue":    {R: 59, G: 105, B: 161, A: 255},
		"magenta": {R: 153, G: 79, B: 130, A: 255},
		"cyan":    {R: 40, G: 131, B: 143, A: 255},
		"white":   {R: 111, G: 118, B: 113, A: 255},
		"pink":    {R: 207, G: 112, B: 151, A: 255},
		"teal":    {R: 36, G: 143, B: 129, A: 255},
	}
	castlePoints = map[string]image.Point{
		"Tripoli":    {X: 205, Y: 285},
		"Acre":       {X: 365, Y: 365},
		"Jerusalem":  {X: 485, Y: 450},
		"Kerak":      {X: 660, Y: 478},
		"Cairo":      {X: 495, Y: 630},
		"Alexandria": {X: 280, Y: 578},
		"Damascus":   {X: 598, Y: 320},
		"Aleppo":     {X: 710, Y: 216},
		"Edessa":     {X: 547, Y: 152},
		"Mosul":      {X: 790, Y: 125},
	}
)

type App struct {
	model          *game.Game
	setup          bool
	name           string
	nameFocused    bool
	startingCastle string
	selectedCastle string
	attackOrigin   string
	moveOrigin     string
	tradeOrigin    string
	notice         string
	lastTick       time.Time
	tickElapsed    time.Duration
	fontSource     *text.GoTextFaceSource
	faces          map[int]*text.GoTextFace
}

func Run() error {
	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return err
	}
	app := &App{
		setup:          true,
		name:           "Ruler",
		startingCastle: "Cairo",
		fontSource:     fontSource,
		faces:          make(map[int]*text.GoTextFace),
	}
	ebiten.SetWindowSize(canvasWidth, canvasHeight)
	ebiten.SetWindowTitle("Apple Conquest | A Levantine strategy game")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(app)
}

func (a *App) Update() error {
	if a.setup {
		a.updateSetup()
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		a.model.Paused = !a.model.Paused
	}
	if a.model != nil && !a.model.Paused && a.model.Outcome == game.Ongoing {
		now := time.Now()
		if !a.lastTick.IsZero() {
			a.tickElapsed += now.Sub(a.lastTick)
		}
		a.lastTick = now
		for a.tickElapsed >= a.model.TickStepDuration() {
			a.model.TickOnce()
			a.tickElapsed -= a.model.TickStepDuration()
		}
	} else {
		a.lastTick = time.Now()
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		a.handleGameClick(x, y)
	}
	return nil
}

func (a *App) Layout(_, _ int) (int, int) {
	return canvasWidth, canvasHeight
}

func (a *App) face(size int) *text.GoTextFace {
	if face := a.faces[size]; face != nil {
		return face
	}
	face := &text.GoTextFace{Source: a.fontSource, Size: float64(size)}
	a.faces[size] = face
	return face
}

func (a *App) drawText(screen *ebiten.Image, value, alignment string, x, y float64, size int, tint color.Color) {
	if alignment == "center" {
		width, _ := text.Measure(value, a.face(size), 0)
		x -= width / 2
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	r, g, b, alpha := tint.RGBA()
	op.ColorScale.Scale(float32(r)/65535, float32(g)/65535, float32(b)/65535, float32(alpha)/65535)
	text.Draw(screen, value, a.face(size), op)
}

func (a *App) drawCentered(screen *ebiten.Image, value string, centerX, y float64, size int, tint color.Color) {
	a.drawText(screen, value, "center", centerX, y, size, tint)
}

func (a *App) drawButton(screen *ebiten.Image, bounds image.Rectangle, label string, fill, textColor color.Color, active bool) {
	if active {
		fill = green
		textColor = panel
	}
	vector.DrawFilledRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), fill, true)
	vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), 1, lineColor, true)
	size := 13
	maxWidth := float64(bounds.Dx() - 12)
	width, height := text.Measure(label, a.face(size), 0)
	for size > 9 && width > maxWidth {
		size--
		width, height = text.Measure(label, a.face(size), 0)
	}
	if width > maxWidth {
		runes := []rune(label)
		for len(runes) > 1 && width > maxWidth {
			runes = runes[:len(runes)-1]
			label = string(runes) + "..."
			width, height = text.Measure(label, a.face(size), 0)
		}
	}
	a.drawText(screen, label, "left", float64(bounds.Min.X)+(float64(bounds.Dx())-width)/2, float64(bounds.Min.Y)+(float64(bounds.Dy())-height)/2, size, textColor)
}

func (a *App) startGame() {
	world, err := game.NewGame(strings.TrimSpace(a.name), "", a.startingCastle, time.Now().UnixNano())
	if err != nil {
		a.notice = err.Error()
		return
	}
	a.model = world
	a.selectedCastle = a.startingCastle
	a.lastTick = time.Now()
	a.tickElapsed = 0
	a.setup = false
}

func (a *App) updateSetup() {
	for _, r := range ebiten.AppendInputChars(nil) {
		if a.nameFocused && unicode.IsPrint(r) && utf8.RuneCountInString(a.name) < 24 {
			a.name += string(r)
		}
	}
	if a.nameFocused && inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(a.name) > 0 {
		_, width := utf8.DecodeLastRuneInString(a.name)
		a.name = a.name[:len(a.name)-width]
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		a.startGame()
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	x, y := ebiten.CursorPosition()
	if image.Pt(x, y).In(image.Rect(78, 270, 466, 329)) {
		a.nameFocused = true
		return
	}
	for index, castleName := range game.CastleNames() {
		column, row := index%2, index/2
		bounds := image.Rect(566+column*353, 267+row*81, 900+column*353, 323+row*81)
		if image.Pt(x, y).In(bounds) {
			a.startingCastle = castleName
			return
		}
	}
	if image.Pt(x, y).In(image.Rect(948, 747, 1288, 809)) {
		a.startGame()
	}
}

func (a *App) handleGameClick(x, y int) {
	if a.model == nil || a.model.Outcome != game.Ongoing {
		return
	}
	if image.Pt(x, y).In(image.Rect(1205, 17, 1328, 57)) {
		a.model.Paused = !a.model.Paused
		return
	}
	for index, preset := range []struct {
		label string
		value string
	}{{"Slower", "slower"}, {"Slow", "slow"}, {"Normal", "normal"}, {"Fast", "fast"}, {"Faster", "faster"}} {
		bounds := image.Rect(820+index*60, 17, 874+index*60, 57)
		if image.Pt(x, y).In(bounds) {
			if err := a.model.SetSpeed(preset.value); err != nil {
				a.notice = err.Error()
			} else {
				a.notice = fmt.Sprintf("Speed set to %s (%d sec/day).", preset.label, a.model.SpeedSeconds)
			}
			return
		}
	}
	for _, name := range a.model.Order {
		point := castlePoints[name]
		distance := math.Hypot(float64(x-point.X), float64(y-point.Y))
		if distance <= 40 {
			if a.moveOrigin != "" && name != a.moveOrigin {
				origin := a.model.Castles[a.moveOrigin]
				if origin == nil {
					a.notice = "Select a valid army source first."
					a.moveOrigin = ""
					return
				}
				amount := origin.Troops
				if amount < 1 {
					a.notice = "This castle has no troops available to move."
					a.moveOrigin = ""
					return
				}
				a.runAction(func() error { return a.model.MoveArmy(a.moveOrigin, name, amount) })
				a.moveOrigin = ""
				return
			}
			if a.tradeOrigin != "" && name != a.tradeOrigin {
				source := a.model.Castles[a.tradeOrigin]
				if source == nil {
					a.notice = "Select a valid source castle for trade."
					a.tradeOrigin = ""
					return
				}
				a.runAction(func() error { return a.model.Trade(a.tradeOrigin, name, 25, 25, 25) })
				a.tradeOrigin = ""
				return
			}
			a.selectedCastle = name
			return
		}
	}
	selected := a.model.Castles[a.selectedCastle]
	if selected == nil {
		return
	}
	if selected.Owner == a.model.PlayerName {
		a.handleOwnedCastleClick(x, y, selected)
	} else {
		a.handleForeignCastleClick(x, y, selected)
	}
}

func (a *App) handleOwnedCastleClick(x, y int, selected *game.Castle) {
	switch {
	case pointIn(x, y, actionRect(0, 0, true)):
		a.attackOrigin = selected.Name
		a.notice = fmt.Sprintf("%s is ready as your attack origin.", selected.Name)
	case pointIn(x, y, actionRect(1, 0, false)):
		a.runAction(func() error { return a.model.Train(selected.Name, "soldiers", 10) })
	case pointIn(x, y, actionRect(1, 1, false)):
		a.runAction(func() error { return a.model.Train(selected.Name, "artillery", 10) })
	case pointIn(x, y, actionRect(1, 2, false)):
		a.runAction(func() error { return a.model.Build(selected.Name, game.Farm) })
	case pointIn(x, y, actionRect(2, 0, false)):
		a.runAction(func() error { return a.model.Build(selected.Name, game.Sawmill) })
	case pointIn(x, y, actionRect(2, 1, false)):
		a.runAction(func() error { return a.model.Build(selected.Name, game.Market) })
	case pointIn(x, y, actionRect(2, 2, false)):
		a.runAction(func() error { return a.model.Build(selected.Name, game.Barracks) })
	case pointIn(x, y, actionRect(3, 0, false)):
		a.runAction(func() error { return a.model.Build(selected.Name, game.Artillery) })
	case pointIn(x, y, actionRect(3, 1, false)):
		a.runAction(func() error { return a.model.Build(selected.Name, game.Forge) })
	case pointIn(x, y, actionRect(3, 2, false)):
		a.runAction(func() error { return a.model.Upgrade(selected.Name) })
	case pointIn(x, y, actionRect(4, 0, false)):
		amount := min(10, selected.Troops)
		if amount == 0 {
			a.notice = "No field troops are available to garrison."
		} else {
			a.runAction(func() error { return a.model.SetGarrison(selected.Name, selected.Garrison+amount) })
		}
	case pointIn(x, y, actionRect(4, 1, false)):
		amount := min(10, selected.Garrison)
		if amount == 0 {
			a.notice = "This castle has no garrison to withdraw."
		} else {
			a.runAction(func() error { return a.model.WithdrawGarrison(selected.Name, amount) })
		}
	case pointIn(x, y, actionRect(4, 2, false)):
		a.moveOrigin = selected.Name
		a.tradeOrigin = ""
		a.notice = fmt.Sprintf("Select a friendly or allied castle to move troops from %s.", selected.Name)
	}
}

func (a *App) handleForeignCastleClick(x, y int, selected *game.Castle) {
	switch {
	case pointIn(x, y, actionRect(0, 0, false)):
		a.runAction(func() error { return a.model.Diplomacy("war", selected.Name) })
	case pointIn(x, y, actionRect(0, 1, false)):
		a.runAction(func() error { return a.model.Diplomacy("truce", selected.Name) })
	case pointIn(x, y, actionRect(1, 0, false)):
		a.runAction(func() error { return a.model.Diplomacy("ally", selected.Name) })
	case pointIn(x, y, actionRect(1, 1, false)):
		a.runAction(func() error { return a.model.Diplomacy("vassal", selected.Name) })
	case pointIn(x, y, actionRect(2, 0, true)) && a.attackOrigin != "":
		origin := a.model.Castles[a.attackOrigin]
		if origin == nil {
			a.notice = "Select a valid attack origin first."
			return
		}
		a.runAction(func() error { return a.model.Attack(a.attackOrigin, selected.Name, origin.Troops/2) })
	case pointIn(x, y, actionRect(3, 0, false)):
		a.runAction(func() error { return a.model.RequestArmy(selected.Name) })
	case pointIn(x, y, actionRect(3, 1, false)):
		a.tradeOrigin = selected.Name
		a.notice = fmt.Sprintf("Select a source castle to trade resources with %s.", selected.Name)
	}
}

func (a *App) runAction(action func() error) {
	if err := action(); err != nil {
		a.notice = err.Error()
		return
	}
	if len(a.model.Events) > 0 {
		a.notice = a.model.Events[len(a.model.Events)-1]
	}
}

func pointIn(x, y int, bounds image.Rectangle) bool {
	return image.Pt(x, y).In(bounds)
}

func containsCastle(values []string, name string) bool {
	for _, value := range values {
		if value == name {
			return true
		}
	}
	return false
}

func actionRect(row, column int, fullWidth bool) image.Rectangle {
	if fullWidth {
		return image.Rect(948, 414+row*44, 1308, 450+row*44)
	}
	return image.Rect(948+column*118, 414+row*44, 1058+column*118, 450+row*44)
}

func (a *App) Draw(screen *ebiten.Image) {
	if a.setup {
		a.drawSetup(screen)
		return
	}
	a.drawGame(screen)
}
