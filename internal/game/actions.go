package game

import (
	"fmt"
	"strings"
)

func (g *Game) Build(castleName string, building Building) error {
	c, err := g.Castle(castleName)
	if err != nil {
		return err
	}
	if c.Owner != g.PlayerName {
		return fmt.Errorf("you can only manage castles you directly control")
	}
	cost, ok := buildingCosts[building]
	if !ok {
		return fmt.Errorf("unknown building %q", building)
	}
	if c.Gold < cost.gold || c.Wood < cost.wood {
		return fmt.Errorf("%s requires %d gold and %d wood", building, cost.gold, cost.wood)
	}
	c.Gold -= cost.gold
	c.Wood -= cost.wood
	c.Buildings[building]++
	g.addEvent(fmt.Sprintf("%s began construction of a %s in %s.", g.PlayerName, building, c.Name))
	return nil
}

func (g *Game) Train(castleName, unit string, amount int) error {
	c, err := g.Castle(castleName)
	if err != nil {
		return err
	}
	if c.Owner != g.PlayerName {
		return fmt.Errorf("you can only train troops in a castle you directly control")
	}
	if amount < 1 || amount > 500 {
		return fmt.Errorf("training amount must be between 1 and 500")
	}
	switch strings.ToLower(unit) {
	case "soldier", "soldiers":
		if c.Buildings[Barracks] == 0 {
			return fmt.Errorf("build a barracks before training soldiers")
		}
		gold, wood, food := amount*2, amount, amount
		if c.Gold < gold || c.Wood < wood || c.Food < food {
			return fmt.Errorf("training %d soldiers requires %d gold, %d wood, and %d food", amount, gold, wood, food)
		}
		c.Gold -= gold
		c.Wood -= wood
		c.Food -= food
		c.Troops += amount
	case "artillery", "siege":
		if c.Buildings[Artillery] == 0 {
			return fmt.Errorf("build an artillery workshop before training artillery")
		}
		gold, wood := amount*8, amount*5
		if c.Gold < gold || c.Wood < wood {
			return fmt.Errorf("training %d artillery requires %d gold and %d wood", amount, gold, wood)
		}
		c.Gold -= gold
		c.Wood -= wood
		c.Artillery += amount
	default:
		return fmt.Errorf("unit must be soldiers or artillery")
	}
	g.addEvent(fmt.Sprintf("%s trained %d %s in %s.", g.PlayerName, amount, unit, c.Name))
	return nil
}

func (g *Game) SetGarrison(castleName string, amount int) error {
	c, err := g.Castle(castleName)
	if err != nil {
		return err
	}
	if c.Owner != g.PlayerName {
		return fmt.Errorf("you can only set garrisons in a castle you directly control")
	}
	if amount < 0 || amount > c.Troops+c.Garrison {
		return fmt.Errorf("garrison must be between 0 and %d troops", c.Troops+c.Garrison)
	}
	c.Troops += c.Garrison - amount
	c.Garrison = amount
	g.addEvent(fmt.Sprintf("%s assigned %d troops to the garrison at %s.", g.PlayerName, amount, c.Name))
	return nil
}

func (g *Game) Upgrade(castleName string) error {
	c, err := g.Castle(castleName)
	if err != nil {
		return err
	}
	if c.Owner != g.PlayerName {
		return fmt.Errorf("you can only upgrade a castle you directly control")
	}
	if c.Buildings[Forge] == 0 {
		return fmt.Errorf("build a forge before upgrading the army")
	}
	if c.Gold < 100 || c.Wood < 40 {
		return fmt.Errorf("an army upgrade requires 100 gold and 40 wood")
	}
	c.Gold -= 100
	c.Wood -= 40
	c.ForgeLevel++
	g.addEvent(fmt.Sprintf("%s upgraded the army at %s to level %d.", g.PlayerName, c.Name, c.ForgeLevel))
	return nil
}

func (g *Game) Diplomacy(action, castleName string) error {
	c, err := g.Castle(castleName)
	if err != nil {
		return err
	}
	ruler := c.Owner
	if ruler == g.PlayerName || g.Vassals[ruler] {
		return fmt.Errorf("%s is already under your control", c.Name)
	}
	current := g.Relations[ruler]
	switch strings.ToLower(action) {
	case "war":
		g.Relations[ruler] = War
		g.addEvent(fmt.Sprintf("%s declared war on %s.", g.PlayerName, ruler))
	case "rival", "rivalry":
		if current == War {
			return fmt.Errorf("you cannot declare a rivalry while at war")
		}
		g.Relations[ruler] = Rivalry
		g.addEvent(fmt.Sprintf("%s declared %s a rival.", g.PlayerName, ruler))
	case "truce":
		if current != War && current != Rivalry {
			return fmt.Errorf("there is no war or rivalry to end with %s", ruler)
		}
		g.Relations[ruler] = Truce
		g.addEvent(fmt.Sprintf("A truce was agreed with %s.", ruler))
	case "ally", "alliance":
		if current == War {
			return fmt.Errorf("an alliance cannot be proposed during war")
		}
		if g.rng.Intn(100) < 65 {
			g.Relations[ruler] = Alliance
			g.addEvent(fmt.Sprintf("%s accepted an alliance with %s.", ruler, g.PlayerName))
		} else {
			g.addEvent(fmt.Sprintf("%s declined the alliance proposal.", ruler))
		}
	case "vassal":
		if current == War {
			return fmt.Errorf("a vassal offer cannot be made during war")
		}
		if g.rulerPower(g.PlayerName) >= g.rulerPower(ruler) || g.rng.Intn(100) < 35 {
			g.Vassals[ruler] = true
			g.Relations[ruler] = Alliance
			g.addEvent(fmt.Sprintf("%s swore fealty to %s.", ruler, g.PlayerName))
		} else {
			g.addEvent(fmt.Sprintf("%s refused to become a vassal.", ruler))
		}
	default:
		return fmt.Errorf("diplomacy action must be war, rival, truce, ally, or vassal")
	}
	g.CheckOutcome()
	return nil
}

func (g *Game) Attack(fromName, targetName string, troops int) error {
	from, err := g.Castle(fromName)
	if err != nil {
		return err
	}
	target, err := g.Castle(targetName)
	if err != nil {
		return err
	}
	if from.Owner != g.PlayerName {
		return fmt.Errorf("attacks must be launched from a castle you directly control")
	}
	if troops < 1 || troops > from.Troops {
		return fmt.Errorf("you have %d field troops available at %s", from.Troops, from.Name)
	}
	if g.PlayerControls(target) {
		return fmt.Errorf("%s is already under your control", target.Name)
	}
	if !contains(from.Neighbors, target.Name) {
		return fmt.Errorf("%s does not border %s", from.Name, target.Name)
	}
	if g.Relations[target.Owner] != War {
		return fmt.Errorf("declare war on %s before attacking", target.Owner)
	}
	g.resolveBattle(from, target, troops)
	g.CheckOutcome()
	return nil
}

func (g *Game) rulerPower(ruler string) int {
	power := 0
	for _, c := range g.Castles {
		if c.Owner == ruler {
			power += c.Troops + c.Garrison + c.Artillery*4
		}
	}
	return power
}

func (g *Game) resolveBattle(attacker, defender *Castle, committed int) {
	attackPower := committed + attacker.Artillery*4 + attacker.ForgeLevel*10
	defensePower := defender.Troops + defender.Garrison + defender.Artillery*4 + defender.ForgeLevel*10
	attacker.Troops -= committed
	if attackPower > defensePower {
		losses := committed / 5
		if losses < 1 {
			losses = 1
		}
		if losses > committed {
			losses = committed
		}
		defender.Owner = attacker.Owner
		defender.Troops = committed - losses
		defender.Garrison = 0
		defender.Artillery /= 2
		g.addEvent(fmt.Sprintf("%s captured %s after defeating its garrison.", attacker.Owner, defender.Name))
	} else {
		losses := committed / 2
		if losses < 1 {
			losses = 1
		}
		attacker.Troops += committed - losses
		g.addEvent(fmt.Sprintf("%s's attack on %s was repelled; %d troops were lost.", attacker.Owner, defender.Name, losses))
	}
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
