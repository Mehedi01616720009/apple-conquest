package terminal

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"apple-conquest/internal/game"
)

func Run(input io.Reader, output io.Writer, currentGame *game.Game) error {
	if _, err := fmt.Fprintln(output, "Apple Conquest | type help for commands"); err != nil {
		return err
	}
	if err := printState(output, currentGame.State()); err != nil {
		return err
	}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		shouldQuit, err := execute(strings.Fields(line), output, currentGame)
		if err != nil {
			if _, writeErr := fmt.Fprintf(output, "Error: %v\n", err); writeErr != nil {
				return writeErr
			}
			continue
		}
		if shouldQuit {
			_, err := fmt.Fprintln(output, "Goodbye.")
			return err
		}
	}
	return scanner.Err()
}

func execute(fields []string, output io.Writer, currentGame *game.Game) (bool, error) {
	command := strings.ToLower(fields[0])
	switch command {
	case "help":
		return false, printHelp(output)
	case "status":
		return false, printState(output, currentGame.State())
	case "quit", "exit":
		return true, nil
	case "end", "turn":
		if len(fields) != 1 {
			return false, fmt.Errorf("usage: end")
		}
		if err := currentGame.AdvanceTurn(); err != nil {
			return false, err
		}
		return false, printState(output, currentGame.State())
	case "build":
		if len(fields) < 2 {
			return false, fmt.Errorf("usage: build <building name>")
		}
		building, ok := parseBuilding(strings.ToLower(strings.Join(fields[1:], " ")))
		if !ok {
			return false, fmt.Errorf("unknown building %q", strings.Join(fields[1:], " "))
		}
		if err := currentGame.Build(building); err != nil {
			return false, err
		}
		return false, printState(output, currentGame.State())
	case "train":
		if len(fields) != 2 {
			return false, fmt.Errorf("usage: train <archer|infantry|cavalry>")
		}
		unit, ok := parseUnit(strings.ToLower(fields[1]))
		if !ok {
			return false, fmt.Errorf("unknown unit %q", fields[1])
		}
		if err := currentGame.Train(unit); err != nil {
			return false, err
		}
		return false, printState(output, currentGame.State())
	case "trade":
		if len(fields) != 4 {
			return false, fmt.Errorf("usage: trade <wood|food|gold> <wood|food|gold> <amount>")
		}
		from, fromOK := parseResource(strings.ToLower(fields[1]))
		to, toOK := parseResource(strings.ToLower(fields[2]))
		amount, amountErr := strconv.Atoi(fields[3])
		if !fromOK || !toOK || amountErr != nil {
			return false, fmt.Errorf("usage: trade <wood|food|gold> <wood|food|gold> <amount>")
		}
		if err := currentGame.Trade(from, to, amount); err != nil {
			return false, err
		}
		return false, printState(output, currentGame.State())
	default:
		return false, fmt.Errorf("unknown command %q; type help for commands", fields[0])
	}
}

func printHelp(output io.Writer) error {
	_, err := fmt.Fprintln(output, `Commands:
  status                         show your settlement
  build <building name>          build a hut, saw mill, wheat field, tax collector, market, or barracks
  train <unit>                   train an archer, infantry, or cavalry (requires a barracks)
  trade <from> <to> <amount>     trade resources at a market
  end                            produce resources, feed civilians, and grow the population
  quit                           leave the game`)
	return err
}

func printState(output io.Writer, state game.State) error {
	if _, err := fmt.Fprintf(output, "\nTurn %d | %s\n", state.Turn, state.Status); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(output, "Resources: wood %d | food %d | gold %d\n", state.Resources[game.Wood], state.Resources[game.Food], state.Resources[game.Gold]); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(output, "Population: %d | Housing: %d\n", state.Population, state.HousingCapacity()); err != nil {
		return err
	}
	if _, err := fmt.Fprint(output, "Buildings:"); err != nil {
		return err
	}
	for _, building := range []game.Building{game.Castle, game.Hut, game.SawMill, game.WheatField, game.TaxCollector, game.Market, game.Barracks} {
		if count := state.Buildings[building]; count > 0 {
			if _, err := fmt.Fprintf(output, " %s %d;", building, count); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(output); err != nil {
		return err
	}
	if _, err := fmt.Fprint(output, "Military:"); err != nil {
		return err
	}
	for _, unit := range []game.Unit{game.Archer, game.Infantry, game.Cavalry} {
		if _, err := fmt.Fprintf(output, " %s %d;", unit, state.Units[unit]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(output); err != nil {
		return err
	}
	if state.Status == game.Won {
		_, err := fmt.Fprintln(output, "Victory: your settlement has met its objective.")
		return err
	}
	if state.Status == game.Lost {
		_, err := fmt.Fprintln(output, "Defeat: your settlement has no civilians left.")
		return err
	}
	return nil
}

func parseBuilding(name string) (game.Building, bool) {
	buildings := map[string]game.Building{
		"castle":        game.Castle,
		"hut":           game.Hut,
		"saw mill":      game.SawMill,
		"sawmill":       game.SawMill,
		"wheat field":   game.WheatField,
		"tax collector": game.TaxCollector,
		"market":        game.Market,
		"barracks":      game.Barracks,
	}
	building, ok := buildings[name]
	return building, ok
}

func parseUnit(name string) (game.Unit, bool) {
	units := map[string]game.Unit{
		"archer":   game.Archer,
		"infantry": game.Infantry,
		"cavalry":  game.Cavalry,
	}
	unit, ok := units[name]
	return unit, ok
}

func parseResource(name string) (game.Resource, bool) {
	resources := map[string]game.Resource{
		"wood": game.Wood,
		"food": game.Food,
		"gold": game.Gold,
	}
	resource, ok := resources[name]
	return resource, ok
}
