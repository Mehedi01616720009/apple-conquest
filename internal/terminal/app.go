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
	"orange":  "\033[38;5;208m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"white":   "\033[37m",
	"pink":    "\033[38;5;205m",
	"teal":    "\033[38;5;37m",
}

func Run(noColor bool) error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	name, err := prompt(scanner, "Ruler name: ")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "\nChoose your starting castle:")
	for index, name := range game.CastleNames() {
		fmt.Fprintf(os.Stdout, "  %02d  %s\n", index+1, name)
	}
	startingCastle, err := prompt(scanner, "Castle number or name: ")
	if err != nil {
		return err
	}

	world, err := game.NewGame(name, "", startingCastle, time.Now().UnixNano())
	if err != nil {
		return err
	}
	app := &App{Game: world, Out: os.Stdout, Color: !noColor && os.Getenv("NO_COLOR") == ""}
	app.eventRead = len(world.Events)
	fmt.Fprintln(app.Out)
	app.printBanner()
	app.RenderMap()
	fmt.Fprintln(app.Out, "Type help to see available commands. Time advances every 12 seconds.")

	lines := make(chan string)
	go func() {
		defer close(lines)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()

	ticker := time.NewTicker(app.Game.TickStepDuration())
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
			previousSpeed := app.Game.SpeedSeconds
			if err := app.Execute(line); err != nil {
				fmt.Fprintf(app.Out, "Error: %v\n", err)
			}
			if previousSpeed != app.Game.SpeedSeconds {
				ticker.Reset(app.Game.TickStepDuration())
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
	case "speed":
		if len(parts) == 1 {
			fmt.Fprintf(a.Out, "Current speed: %s (%d sec/day)\n", speedLabel(a.Game.SpeedSeconds), a.Game.SpeedSeconds)
			return nil
		}
		if len(parts) != 2 {
			return fmt.Errorf("usage: speed <slower|slow|normal|fast|faster>")
		}
		if err := a.Game.SetSpeed(parts[1]); err != nil {
			return err
		}
		fmt.Fprintf(a.Out, "Speed set to %s (%d sec/day).\n", speedLabel(a.Game.SpeedSeconds), a.Game.SpeedSeconds)
		return nil
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
	case "withdraw":
		if len(parts) != 3 {
			return fmt.Errorf("usage: withdraw <castle#> <troops>")
		}
		amount, err := parseAmount(parts[2])
		if err != nil {
			return err
		}
		return a.Game.WithdrawGarrison(parts[1], amount)
	case "move":
		if len(parts) != 4 {
			return fmt.Errorf("usage: move <from#> <to#> <troops>")
		}
		amount, err := parseAmount(parts[3])
		if err != nil {
			return err
		}
		return a.Game.MoveArmy(parts[1], parts[2], amount)
	case "request":
		if len(parts) != 2 {
			return fmt.Errorf("usage: request <ally-or-vassal-castle#>")
		}
		return a.Game.RequestArmy(parts[1])
	case "trade":
		if len(parts) != 6 {
			return fmt.Errorf("usage: trade <source#> <target#> <gold> <wood> <food>")
		}
		gold, err := parseAmount(parts[3])
		if err != nil {
			return err
		}
		wood, err := parseAmount(parts[4])
		if err != nil {
			return err
		}
		food, err := parseAmount(parts[5])
		if err != nil {
			return err
		}
		return a.Game.Trade(parts[1], parts[2], gold, wood, food)
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
	fmt.Fprintln(a.Out, "  speed <slower|slow|normal|fast|faster>  set day pace")
	fmt.Fprintln(a.Out, "  build <castle#> <building>               construct an economy or army building")
	fmt.Fprintln(a.Out, "  train <castle#> <soldiers|artillery> <n> raise units")
	fmt.Fprintln(a.Out, "  garrison <castle#> <n>                   assign defensive troops")
	fmt.Fprintln(a.Out, "  withdraw <castle#> <n>                   move garrison troops to the field army")
	fmt.Fprintln(a.Out, "  move <from#> <to#> <troops>              send troops to a friendly or allied castle")
	fmt.Fprintln(a.Out, "  request <castle#>                        request 15% of an adjacent ally/vassal army (50% chance)")
	fmt.Fprintln(a.Out, "  trade <source#> <target#> <gold> <wood> <food>  exchange resources with allies or vassals")
	fmt.Fprintln(a.Out, "  upgrade <castle#>                        improve troops at a forge")
	fmt.Fprintln(a.Out, "  diplomacy <war|rival|truce|ally|vassal> <castle#>")
	fmt.Fprintln(a.Out, "  attack <from#> <target#> <troops>        attack an adjacent enemy castle after 2-day march and 1-day battle")
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

func speedLabel(seconds int) string {
	switch seconds {
	case 25:
		return "slower"
	case 20:
		return "slow"
	case 15:
		return "normal"
	case 10:
		return "fast"
	case 5:
		return "faster"
	default:
		return "normal"
	}
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
