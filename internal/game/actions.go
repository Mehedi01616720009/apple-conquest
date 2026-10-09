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

func (g *Game) WithdrawGarrison(castleName string, amount int) error {
	c, err := g.Castle(castleName)
	if err != nil {
		return err
	}
	if c.Owner != g.PlayerName {
		return fmt.Errorf("you can only withdraw troops from a castle you directly control")
	}
	if amount < 1 || amount > c.Garrison {
		return fmt.Errorf("withdrawal must be between 1 and %d garrison troops", c.Garrison)
	}
	c.Garrison -= amount
	c.Troops += amount
	g.addEvent(fmt.Sprintf("%s withdrew %d troops from the garrison at %s.", g.PlayerName, amount, c.Name))
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
		if g.acceptDiplomacyOffer("truce", ruler) {
			g.Relations[ruler] = Truce
			g.addEvent(fmt.Sprintf("%s accepted a truce with %s.", ruler, g.PlayerName))
		} else {
			g.Relations[ruler] = Peace
			g.addEvent(fmt.Sprintf("%s rejected the truce offer.", ruler))
		}
	case "ally", "alliance":
		if current == War {
			return fmt.Errorf("an alliance cannot be proposed during war")
		}
		if g.acceptDiplomacyOffer("alliance", ruler) {
			g.Relations[ruler] = Alliance
			g.addEvent(fmt.Sprintf("%s accepted an alliance with %s.", ruler, g.PlayerName))
		} else {
			g.Relations[ruler] = Rivalry
			g.addEvent(fmt.Sprintf("%s declined the alliance proposal.", ruler))
		}
	case "vassal":
		if current == War {
			return fmt.Errorf("a vassal offer cannot be made during war")
		}
		if g.acceptDiplomacyOffer("vassal", ruler) {
			g.Vassals[ruler] = true
			g.VassalParents[ruler] = g.PlayerName
			g.Relations[ruler] = Alliance
			g.addEvent(fmt.Sprintf("%s swore fealty to %s.", ruler, g.PlayerName))
		} else {
			g.Relations[ruler] = Rivalry
			g.addEvent(fmt.Sprintf("%s refused to become a vassal.", ruler))
		}
	default:
		return fmt.Errorf("diplomacy action must be war, rival, truce, ally, or vassal")
	}
	g.CheckOutcome()
	return nil
}

func (g *Game) acceptDiplomacyOffer(action, ruler string) bool {
	myResources := g.resourceSumFor(g.PlayerName)
	targetResources := g.resourceSumFor(ruler)
	myArmy := g.armyTotalFor(g.PlayerName)
	targetArmy := g.armyTotalFor(ruler)

	if targetResources <= 0 {
		targetResources = 1
	}
	if targetArmy <= 0 {
		targetArmy = 1
	}

	switch action {
	case "alliance":
		if myResources <= targetResources {
			return false
		}
		return withinTolerance(myArmy, targetArmy, 0.20)
	case "vassal":
		if myResources < int(float64(targetResources)*1.6) {
			return false
		}
		return myArmy >= int(float64(targetArmy)*1.5)
	case "truce":
		if myResources >= int(float64(targetResources)*1.6) && myArmy >= int(float64(targetArmy)*1.6) {
			return true
		}
		if myResources >= int(float64(targetResources)*1.35) && myResources < int(float64(targetResources)*1.6) &&
			myArmy >= int(float64(targetArmy)*1.4) && myArmy < int(float64(targetArmy)*1.6) {
			return g.rng.Float64() < 0.5
		}
		return false
	default:
		return false
	}
}

func (g *Game) resourceSumFor(owner string) int {
	total := 0
	for _, c := range g.Castles {
		if c.Owner == owner {
			total += c.Gold + c.Wood + c.Food
		}
	}
	return total
}

func (g *Game) armyTotalFor(owner string) int {
	total := 0
	for _, c := range g.Castles {
		if c.Owner == owner {
			total += c.Troops + c.Garrison
		}
	}
	return total
}

func withinTolerance(valueA, valueB int, tolerance float64) bool {
	if valueA == valueB {
		return true
	}
	low, high := valueA, valueB
	if low > high {
		low, high = high, low
	}
	if high == 0 {
		return true
	}
	return float64(high-low) <= tolerance*float64(high)
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
	if !g.PlayerControls(from) {
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
	g.Orders = append(g.Orders, ArmyOrder{Type: "attack", From: from.Name, To: target.Name, Owner: from.Owner, Troops: troops, Artillery: from.Artillery, TravelDays: 2, HoldDays: 1})
	g.addEvent(fmt.Sprintf("%s ordered %d troops from %s to march toward %s; they will arrive in 2 days and attack for 1 day.", from.Owner, troops, from.Name, target.Name))
	g.CheckOutcome()
	return nil
}

func (g *Game) MoveArmy(fromName, toName string, troops int) error {
	from, err := g.Castle(fromName)
	if err != nil {
		return err
	}
	to, err := g.Castle(toName)
	if err != nil {
		return err
	}
	if from.Name == to.Name {
		return fmt.Errorf("armies cannot be moved to the same castle")
	}
	if !g.PlayerControls(from) {
		return fmt.Errorf("you can only move troops from a castle you control")
	}
	if troops < 1 || troops > from.Troops {
		return fmt.Errorf("you have %d field troops available at %s", from.Troops, from.Name)
	}
	if !(to.Owner == from.Owner || to.Owner == g.PlayerName || g.Vassals[to.Owner] || g.Relations[to.Owner] == Alliance || g.Relations[to.Owner] == Truce) {
		return fmt.Errorf("you can only move troops to a castle in your realm, an ally, or a vassal")
	}
	if from.Owner != to.Owner && !contains(from.Neighbors, to.Name) {
		return fmt.Errorf("%s does not border %s", from.Name, to.Name)
	}
	g.Orders = append(g.Orders, ArmyOrder{Type: "move", From: from.Name, To: to.Name, Owner: from.Owner, Troops: troops, TravelDays: 2})
	g.addEvent(fmt.Sprintf("%s moved %d troops from %s to %s; arrival in 2 days.", from.Owner, troops, from.Name, to.Name))
	return nil
}

func (g *Game) RequestArmy(castleName string) error {
	c, err := g.Castle(castleName)
	if err != nil {
		return err
	}
	if !g.Vassals[c.Owner] && g.Relations[c.Owner] != Alliance {
		return fmt.Errorf("you can only request reinforcements from an ally or vassal")
	}
	var destination *Castle
	for _, neighborName := range c.Neighbors {
		neighbor := g.Castles[neighborName]
		if neighbor != nil && neighbor.Owner == g.PlayerName {
			destination = neighbor
			break
		}
	}
	if destination == nil {
		return fmt.Errorf("%s has no neighboring castle under your direct control", c.Name)
	}
	if g.rng.Intn(100) < 50 {
		g.addEvent(fmt.Sprintf("%s refused the request for reinforcements.", c.Owner))
		return nil
	}
	troops := (c.Troops + c.Garrison) * 15 / 100
	if troops < 1 {
		g.addEvent(fmt.Sprintf("%s agreed, but has no 15%% army share available to send.", c.Owner))
		return nil
	}
	fromField := min(troops, c.Troops)
	c.Troops -= fromField
	c.Garrison -= troops - fromField
	g.Orders = append(g.Orders, ArmyOrder{Type: "reinforcement", From: c.Name, To: destination.Name, Owner: c.Owner, Troops: troops, TravelDays: 2})
	g.addEvent(fmt.Sprintf("%s agreed to send %d troops to %s; arrival in 2 days.", c.Owner, troops, destination.Name))
	return nil
}

func (g *Game) Trade(sourceName, targetName string, gold, wood, food int) error {
	source, err := g.Castle(sourceName)
	if err != nil {
		return err
	}
	target, err := g.Castle(targetName)
	if err != nil {
		return err
	}
	if gold < 0 || wood < 0 || food < 0 {
		return fmt.Errorf("trade quantities must be non-negative")
	}
	if !g.PlayerControls(source) {
		return fmt.Errorf("you can only trade from a castle you control")
	}
	if !(target.Owner == source.Owner || target.Owner == g.PlayerName || g.Vassals[target.Owner] || g.Relations[target.Owner] == Alliance || g.Relations[target.Owner] == Truce) {
		return fmt.Errorf("trade is only allowed within your realm, with a vassal, or with an ally")
	}
	if source.Gold < gold || source.Wood < wood || source.Food < food {
		return fmt.Errorf("%s does not have enough resources to trade", source.Name)
	}
	source.Gold -= gold
	source.Wood -= wood
	source.Food -= food
	target.Gold += gold
	target.Wood += wood
	target.Food += food
	g.addEvent(fmt.Sprintf("%s traded %d gold, %d wood, and %d food with %s.", source.Name, gold, wood, food, target.Name))
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
	defendingTroops := defender.Troops + defender.Garrison
	attackPower := armyPower(committed, attacker.Artillery, attacker.ForgeLevel)
	defensePower := armyPower(defendingTroops, defender.Artillery, defender.ForgeLevel)
	attackerLosses := troopsLost(defensePower, attacker.ForgeLevel, committed)
	defenderLosses := troopsLost(attackPower, defender.ForgeLevel, defendingTroops)
	attacker.Troops -= committed
	if attackPower > defensePower {
		applyDefenderLosses(defender, defenderLosses)
		defender.Owner = attacker.Owner
		defender.Troops += committed - attackerLosses
		defender.Artillery /= 2
		g.addEvent(fmt.Sprintf("%s captured %s: attackers lost %d troops; defenders lost %d.", attacker.Owner, defender.Name, attackerLosses, defenderLosses))
	} else {
		applyDefenderLosses(defender, defenderLosses)
		attacker.Troops += committed - attackerLosses
		g.addEvent(fmt.Sprintf("%s's attack on %s was repelled: attackers lost %d troops; defenders lost %d.", attacker.Owner, defender.Name, attackerLosses, defenderLosses))
	}
}

func armyPower(troops, artillery, forgeLevel int) int64 {
	if forgeLevel < 0 {
		forgeLevel = 0
	}
	if forgeLevel > 30 {
		forgeLevel = 30
	}
	unitPower := int64(1) << forgeLevel
	return int64(troops)*unitPower + int64(artillery)*4
}

func troopsLost(enemyPower int64, forgeLevel, available int) int {
	if enemyPower <= 0 || available <= 0 {
		return 0
	}
	if forgeLevel < 0 {
		forgeLevel = 0
	}
	if forgeLevel > 30 {
		forgeLevel = 30
	}
	unitPower := int64(1) << forgeLevel
	losses := 1 + (enemyPower-1)/unitPower
	if losses >= int64(available) {
		return available
	}
	return int(losses)
}

func applyDefenderLosses(defender *Castle, losses int) {
	garrisonLosses := min(losses, defender.Garrison)
	defender.Garrison -= garrisonLosses
	defender.Troops -= losses - garrisonLosses
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
