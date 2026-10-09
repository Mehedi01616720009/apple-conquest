# Apple Conquest

Apple Conquest is a real-time, single-player strategy game with a terminal client and a minimal Ebitengine desktop interface. Choose a ruler, color, and castle, then build your economy, raise troops, negotiate with neighboring rulers, and bring the ten-castle realm under your control.

## Requirements

- Go 1.22 or later
- A terminal for the command-line client
- Linux X11/OpenGL development libraries for the native desktop client

The desktop client uses Ebitengine. On Debian or Ubuntu, install its native build prerequisites with `sudo apt install gcc libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev`. The terminal client needs no graphics libraries.

## Run

From the repository root:

```sh
go run ./cmd/conquest
```

To disable ANSI colors:

```sh
go run ./cmd/conquest --no-color
```

You can also set `NO_COLOR=1` to disable colors:

```sh
NO_COLOR=1 go run ./cmd/conquest
```

At startup, enter a ruler name, choose one of `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, or `white`, then select a castle by its number or name. The color defaults to green if left blank. The player takes over the selected castle; the other nine begin under their listed AI rulers.

### Desktop Client

Launch the graphical version with:

```sh
go run ./cmd/conquest-desktop
```

Choose a ruler, banner color, and starting castle in the setup window. In the campaign, click a castle to inspect it and use the action buttons in the right panel. Click an owned castle's attack-origin button, then select an adjacent enemy and press **ATTACK**. The simulation advances in real time; use **PAUSE** or Space to pause and resume.

## Interface

The map is a regional schematic. `P` marks player-held castles, `V` marks castles held by vassals, and `A` marks independent AI rulers. The neighbor list indicates which castles can be attacked from each location.

```text
APPLE CONQUEST
A real-time struggle for the castles of the Levant
------------------------------------------------

REGIONAL MAP ------------------------------------
   +------------------------------------+  +------------------------------------+
   | 08  Edessa                         |  | 09  Aleppo                         |
   |                                    |  |                                    |
   | Ruler: Joscelin III                |  | Ruler: Emir Muzaffar Uddin Gokbori |
   | Control: AI                        |  | Control: AI                        |
   | Relation: peace                    |  | Relation: peace                    |
   |                                    |  |                                    |
   +------------------------------------+  +------------------------------------+

   +------------------------------------+  +------------------------------------+
   | 01  Cairo                          |  | 02  Alexandria                     |
   |                                    |  |                                    |
   | Ruler: Amina                       |  | Ruler: Emir Kazi Fadil             |
   | Control: PLAYER                    |  | Control: AI                        |
   | Relation: peace                    |  | Relation: peace                    |
   |                                    |  |                                    |
   +------------------------------------+  +------------------------------------+

P = player   V = vassal   AI = independent ruler
NEIGHBORS ----------------------------------------
01 Cairo        -> Alexandria, Damascus, Jerusalem
02 Alexandria   -> Cairo
...

> status

REALM STATUS ------------------------------------
Ruler: Amina | Day: 0 | Time: running
-------------------------------------------------
01 Cairo        | Gold  150 Wood   90 Food  130 | Pop  520
   Army level 1: 100 field, 0 garrison | Artillery 0
   Buildings: farm x1, house x1, sawmill x1, market x1, barracks x1
Vassals: 0 | Outcome: ongoing

>
```

The world advances every second, including while you type commands. Ten simulation ticks are displayed as one game day. Use `pause` to stop the simulation and `resume` to continue it. Periodic summaries and game events appear in the terminal; `events` shows recent events again.

## Commands

Castle arguments accept the number shown in the map/list or the castle name. Use numbers for concise commands.

| Command                                                 | Effect                                                                             |
| ------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `help`                                                  | Show the in-game command list.                                                     |
| `map`                                                   | Show the regional map and castle neighbors.                                        |
| `status`                                                | Show resources, population, army, buildings, and vassals in your castles.          |
| `castles`                                               | List each castle's ruler, control status, and diplomatic relation.                 |
| `events`                                                | Show the latest game events.                                                       |
| `pause` / `resume`                                      | Stop or restart real-time simulation ticks.                                        |
| `build <castle#> <building>`                            | Build a `farm`, `house`, `sawmill`, `market`, `barracks`, `artillery`, or `forge`. |
| `train <castle#> <soldiers\|artillery> <amount>`        | Recruit soldiers or artillery.                                                     |
| `garrison <castle#> <amount>`                           | Set how many troops defend a castle; the rest remain available in its field army.  |
| `withdraw <castle#> <amount>`                           | Move troops from a castle's garrison into its field army.                          |
| `upgrade <castle#>`                                     | Spend resources at a forge to improve the castle's army.                           |
| `diplomacy <war\|rival\|truce\|ally\|vassal> <castle#>` | Change relations with the ruler of the target castle or make an offer.             |
| `attack <from#> <target#> <troops>`                     | Attack an adjacent castle after declaring war on its ruler.                        |
| `quit` / `exit`                                         | End the current campaign session.                                                  |

Examples, assuming Cairo is castle 1 and Jerusalem is castle 5:

```text
build 1 farm
train 1 soldiers 25
garrison 1 40
withdraw 1 15
diplomacy war 5
attack 1 5 100
```

## Economy and Military

Buildings are constructed immediately when paid for. Each building costs gold and wood:

| Building  | Gold | Wood | Purpose                        |
| --------- | ---: | ---: | ------------------------------ |
| Farm      |   35 |   12 | Increases food production.     |
| House     |   45 |   18 | Increases population capacity. |
| Sawmill   |   40 |   10 | Increases wood production.     |
| Market    |   55 |   15 | Increases gold income.         |
| Barracks  |   70 |   30 | Required to train soldiers.    |
| Artillery |  100 |   45 | Required to train artillery.   |
| Forge     |   85 |   35 | Required for army upgrades.    |

Training each soldier costs 2 gold, 1 wood, and 1 food. Training each artillery unit costs 8 gold and 5 wood. An army upgrade costs 100 gold and 40 wood. Resources and population change with each simulation tick; troops and artillery also have ongoing upkeep.

Attacks are resolved immediately rather than traveling over time. Attacks use only field troops at the source; its garrison stays to defend that castle. A defending castle fights with both field troops and garrison, with garrison troops taking casualties first. Use `withdraw <castle#> <amount>` to move garrison troops back to the field army before an attack.

Each forge upgrade raises the army level by one, starting at level 1, and each level doubles a soldier's combat strength. Casualties scale with the level difference. For example, a level 1 force of 100 attacking 120 level 1 defenders loses all 100 while inflicting 100 casualties, leaving 20 defenders. A level 2 force of 100 against 120 level 1 defenders wins, loses 60 troops, and captures the castle with 40 survivors.

## Diplomacy and Campaign

- **War:** Required before attacking that ruler's castles.
- **Rivalry:** Marks hostile relations, but attacks still require a declaration of war.
- **Truce:** Ends an existing war or rivalry.
- **Alliance:** A proposal may be accepted or declined by the AI ruler.
- **Vassal:** A ruler may swear fealty based on relative military strength and chance. Their castles count toward your realm while they remain your vassal.

Win by controlling all ten castles directly or through vassals. You lose if you have no directly controlled castles. AI rulers develop their castles, recruit troops, may declare war when bordering your realm, and can attack during war. Diplomacy and AI actions are currently a lightweight first implementation rather than a complete grand-strategy simulation.

## Castles and Rulers

| Castle     | Starting ruler              |
| ---------- | --------------------------- |
| Cairo      | Sultan Salahuddin Ayyubi    |
| Alexandria | Emir Kazi Fadil             |
| Kerak      | Lord Reynald De Chatillon   |
| Damascus   | Sultan Nur Al Din Zengi     |
| Jerusalem  | King Baldwin IV             |
| Tripoli    | Lord Reymond III            |
| Acre       | Baron Balian of Ibelin      |
| Edessa     | Joscelin III                |
| Aleppo     | Emir Muzaffar Uddin Gokbori |
| Mosul      | Emir Saifuddin Zengi        |

## Development

Run the test suite and build all packages with:

```sh
go test ./...
go build ./...
```
