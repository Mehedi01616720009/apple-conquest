package terminal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"apple-conquest/internal/game"
)

type App struct {
	Game      *game.Game
	Out       io.Writer
	Color     bool
	eventRead int
}

var colorCodes = map[string]string{
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"white":   "\033[37m",
}

func Run(noColor bool) error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	name, err := prompt(scanner, "Ruler name: ")
	if err != nil {
		return err
	}
	color, err := prompt(scanner, "Color (red/green/yellow/blue/magenta/cyan/white) [green]: ")
	if err != nil {
		return err
	}
	color = strings.ToLower(strings.TrimSpace(color))
	if color == "" {
		color = "green"
	}
	if _, ok := colorCodes[color]; !ok {
		return fmt.Errorf("unsupported color %q", color)
	}
	fmt.Fprintln(os.Stdout, "\nChoose your starting castle:")
	for index, name := range game.CastleNames() {
		fmt.Fprintf(os.Stdout, "  %02d  %s\n", index+1, name)
	}
	startingCastle, err := prompt(scanner, "Castle number or name: ")
	if err != nil {
		return err
	}

	world, err := game.NewGame(name, color, startingCastle, time.Now().UnixNano())
	if err != nil {
		return err
	}
	app := &App{Game: world, Out: os.Stdout, Color: !noColor && os.Getenv("NO_COLOR") == ""}
	app.eventRead = len(world.Events)
	fmt.Fprintln(app.Out)
	app.printBanner()
	app.RenderMap()
	fmt.Fprintln(app.Out, "Type help to see available commands. Time advances once per second.")

	lines := make(chan string)
	go func() {
		defer close(lines)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	fmt.Fprint(app.Out, "\n> ")
	for {
		select {
		case <-ticker.C:
			app.Game.TickOnce()
			app.printNewEvents()
			if app.Game.Tick > 0 && app.Game.Tick%10 == 0 {
				app.printDaySummary()
			}
			if app.Game.Outcome != game.Ongoing {
				app.printOutcome()
				return nil
			}
		case line, ok := <-lines:
			if !ok {
				return nil
			}
			if strings.EqualFold(strings.TrimSpace(line), "quit") || strings.EqualFold(strings.TrimSpace(line), "exit") {
				fmt.Fprintln(app.Out, "The campaign has been paused. Farewell, ruler.")
				return nil
			}
			if err := app.Execute(line); err != nil {
				fmt.Fprintf(app.Out, "Error: %v\n", err)
			}
			app.printNewEvents()
			if app.Game.Outcome != game.Ongoing {
				app.printOutcome()
				return nil
			}
			fmt.Fprint(app.Out, "\n> ")
		}
	}
}

func prompt(scanner *bufio.Scanner, label string) (string, error) {
	fmt.Fprint(os.Stdout, label)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(scanner.Text()), nil
}

func (a *App) Execute(line string) error {
	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) == 0 {
		return nil
	}
	switch strings.ToLower(parts[0]) {
	case "help":
		a.printHelp()
	case "map":
		a.RenderMap()
	case "status":
		a.RenderStatus()
	case "castles":
		a.RenderCastles()
	case "events":
		a.RenderEvents()
	case "pause":
		a.Game.Paused = true
		fmt.Fprintln(a.Out, "Time paused.")
	case "resume":
		a.Game.Paused = false
		fmt.Fprintln(a.Out, "Time resumed.")
	case "build":
		if len(parts) != 3 {
			return fmt.Errorf("usage: build <castle#> <farm|house|sawmill|market|barracks|artillery|forge>")
		}
		return a.Game.Build(parts[1], game.Building(strings.ToLower(parts[2])))
	case "train":
		if len(parts) != 4 {
			return fmt.Errorf("usage: train <castle#> <soldiers|artillery> <amount>")
		}
		amount, err := parseAmount(parts[3])
		if err != nil {
			return err
		}
		return a.Game.Train(parts[1], parts[2], amount)
	case "garrison":
		if len(parts) != 3 {
			return fmt.Errorf("usage: garrison <castle#> <troops>")
		}
		amount, err := parseAmount(parts[2])
		if err != nil {
			return err
		}
		return a.Game.SetGarrison(parts[1], amount)
	case "upgrade":
		if len(parts) != 2 {
			return fmt.Errorf("usage: upgrade <castle#>")
		}
		return a.Game.Upgrade(parts[1])
	case "diplomacy":
		if len(parts) != 3 {
			return fmt.Errorf("usage: diplomacy <war|rival|truce|ally|vassal> <castle#>")
		}
		return a.Game.Diplomacy(parts[1], parts[2])
	case "attack":
		if len(parts) != 4 {
			return fmt.Errorf("usage: attack <from#> <target#> <troops>")
		}
		amount, err := parseAmount(parts[3])
		if err != nil {
			return err
		}
		return a.Game.Attack(parts[1], parts[2], amount)
	default:
		return fmt.Errorf("unknown command %q; type help", parts[0])
	}
	return nil
}

func (a *App) printBanner() {
	title := "APPLE CONQUEST"
	if a.Color {
		fmt.Fprintf(a.Out, "\033[1;33m%s\033[0m\n", title)
	} else {
		fmt.Fprintln(a.Out, title)
	}
	fmt.Fprintln(a.Out, "A real-time struggle for the castles of the Levant")
	fmt.Fprintln(a.Out, "------------------------------------------------")
}

func (a *App) printHelp() {
	fmt.Fprintln(a.Out, "Commands ----------------------------------------")
	fmt.Fprintln(a.Out, "  map                                      regional castle map")
	fmt.Fprintln(a.Out, "  status | castles | events                inspect your realm")
	fmt.Fprintln(a.Out, "  pause | resume                           control simulation time")
	fmt.Fprintln(a.Out, "  build <castle#> <building>               construct an economy or army building")
	fmt.Fprintln(a.Out, "  train <castle#> <soldiers|artillery> <n> raise units")
	fmt.Fprintln(a.Out, "  garrison <castle#> <n>                   assign defensive troops")
	fmt.Fprintln(a.Out, "  upgrade <castle#>                        improve troops at a forge")
	fmt.Fprintln(a.Out, "  diplomacy <war|rival|truce|ally|vassal> <castle#>")
	fmt.Fprintln(a.Out, "  attack <from#> <target#> <troops>        attack an adjacent enemy castle")
	fmt.Fprintln(a.Out, "  quit                                     end this session")
	fmt.Fprintln(a.Out, "Use castle numbers shown by map/castles. Time keeps advancing while commands are entered.")
}

func (a *App) printNewEvents() {
	for a.eventRead < len(a.Game.Events) {
		fmt.Fprintf(a.Out, "\n  * %s", a.Game.Events[a.eventRead])
		a.eventRead++
	}
}

func (a *App) printDaySummary() {
	owned := 0
	for _, c := range a.Game.Castles {
		if c.Owner == a.Game.PlayerName {
			owned++
		}
	}
	fmt.Fprintf(a.Out, "\nDay %d | Direct castles: %d | Player army: %d troops | Gold: %d\n", a.Game.Tick/10, owned, a.playerTroops(), a.playerGold())
}

func (a *App) playerTroops() int {
	total := 0
	for _, c := range a.Game.Castles {
		if c.Owner == a.Game.PlayerName {
			total += c.Troops + c.Garrison
		}
	}
	return total
}

func (a *App) playerGold() int {
	total := 0
	for _, c := range a.Game.Castles {
		if c.Owner == a.Game.PlayerName {
			total += c.Gold
		}
	}
	return total
}

func (a *App) printOutcome() {
	switch a.Game.Outcome {
	case game.Victory:
		fmt.Fprintf(a.Out, "\n%s has united all ten castles. Victory!\n", a.Game.PlayerName)
	case game.Defeat:
		fmt.Fprintln(a.Out, "\nYour last castle has fallen. The campaign is over.")
	}
}

func parseAmount(value string) (int, error) {
	amount, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%q is not a number", value)
	}
	return amount, nil
}
