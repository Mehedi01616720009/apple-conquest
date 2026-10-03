# Apple Conquest

A small, offline, turn-based terminal strategy game written in Go. The first milestone focuses on settlement economy, population, and training an army. Buildings are counted rather than placed on a map; combat and enemy AI are not implemented yet.

## Run

Requires Go 1.22 or newer.

```sh
go run ./cmd/conquest
```

Run the automated tests with:

```sh
go test ./...
```

## Objective

Reach 8 civilians and train at least one Archer, Infantry, and Cavalry. Civilians consume 1 food each at the end of every turn. If the settlement cannot feed its population, one civilian is lost per missing food; the game ends in defeat when the population reaches zero.

The Castle and one Hut are provided at the start. A Hut houses 5 civilians. At the end of a fully fed turn, the population grows by one if housing is available. Military units do not consume food after training.

One route to the objective is to build a Wheat Field and Saw Mill, advance two turns, build a Tax Collector, advance three turns, build another Hut, and advance four turns. Then build a Barracks and train each unit type. This is a starter balance example, not the only valid strategy.

## Commands

- `status` shows resources, population, buildings, and units.
- `build <building name>` builds a Hut, Saw Mill, Wheat Field, Tax Collector, Market, or Barracks. The Castle is unique and cannot be rebuilt.
- `train <archer|infantry|cavalry>` trains a unit; a Barracks is required.
- `trade <from> <to> <amount>` trades resources at a Market.
- `end` produces resources, feeds civilians, and grows the population.
- `help` displays commands; `quit` exits.

Commands are case-insensitive. Building names with spaces can be entered as written, for example `build saw mill`.

## Starter Rules

Starting resources are 30 wood, 30 food, and 15 gold, with 3 civilians, one Castle, and one Hut. Production and income are applied before food upkeep each turn.

| Building      | Cost             | Effect per turn         |
| ------------- | ---------------- | ----------------------- |
| Hut           | 10 wood          | Adds 5 housing          |
| Saw Mill      | 12 wood          | Produces 4 wood         |
| Wheat Field   | 12 wood          | Produces 5 food         |
| Tax Collector | 12 wood          | Produces 3 gold         |
| Market        | 15 wood          | Enables resource trades |
| Barracks      | 20 wood, 10 gold | Enables unit training   |

| Unit     | Training cost  |
| -------- | -------------- |
| Archer   | 2 food, 2 gold |
| Infantry | 3 food, 3 gold |
| Cavalry  | 5 food, 6 gold |

Market rates are fixed: 2 wood for 1 gold, 3 food for 1 gold, 1 gold for 2 wood, or 1 gold for 3 food. Trade input amounts must match the rate exactly. Balance values are centralized in `internal/game/definitions.go` and are intended to be tuned through playtesting.
