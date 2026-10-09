package game

import (
	"fmt"
	"strings"
)

func (g *Game) TickOnce() {
	if g.Paused || g.Outcome != Ongoing {
		return
	}
	g.Tick++
	for _, name := range g.Order {
		c := g.Castles[name]
		g.advanceEconomy(c)
	}
	if g.Tick%5 == 0 {
		for _, name := range g.Order {
			c := g.Castles[name]
			if c.Owner == g.PlayerName || g.Vassals[c.Owner] {
				continue
			}
			if activity := g.aiDevelop(c); activity != "" {
				g.aiActivity = append(g.aiActivity, activity)
			}
			g.aiConsiderWar(c)
		}
	}
	if g.Tick%10 == 0 {
		if len(g.aiActivity) == 0 {
			g.addEvent(fmt.Sprintf("Day %d: the realm continues to develop.", g.Tick/10))
		} else {
			const maxDetails = 3
			details := g.aiActivity
			if len(details) > maxDetails {
				details = details[:maxDetails]
			}
			message := fmt.Sprintf("Day %d: AI activity (%d actions): %s", g.Tick/10, len(g.aiActivity), joinActivity(details))
			if extra := len(g.aiActivity) - len(details); extra > 0 {
				message += fmt.Sprintf("; and %d more", extra)
			}
			g.addEvent(message)
			g.aiActivity = nil
		}
	}
	g.CheckOutcome()
}

func (g *Game) advanceEconomy(c *Castle) {
	troopCount := c.Troops + c.Garrison
	c.Food += c.Buildings[Farm]*8 + c.Population/120 - troopCount/20
	c.Wood += c.Buildings[Sawmill] * 5
	c.Gold += 3 + c.Buildings[Market]*6 + c.Population/30 - troopCount/10 - c.Artillery*2
	if c.Food < 0 {
		c.Food = 0
	}
	if c.Gold < 0 {
		c.Gold = 0
	}
	if c.Wood < 0 {
		c.Wood = 0
	}
	capacity := 500 + c.Buildings[House]*250
	if c.Food > 0 && c.Population < capacity {
		c.Population++
	} else if c.Food == 0 && g.Tick%10 == 0 && c.Population > 100 {
		c.Population -= 5
	}
}

func (g *Game) aiDevelop(c *Castle) string {
	if g.Relations[c.Owner] == War && g.tryAIAttack(c) {
		return ""
	}
	if c.Buildings[Barracks] > 0 && c.Troops < 180 && c.Gold >= 30 && c.Wood >= 15 && c.Food >= 10 {
		c.Gold -= 30
		c.Wood -= 15
		c.Food -= 10
		c.Troops += 10
		return fmt.Sprintf("%s recruited 10 troops", c.Name)
	} else if c.Buildings[Artillery] == 0 && c.Troops >= 180 && canAfford(c, Artillery) {
		g.aiBuild(c, Artillery)
		return fmt.Sprintf("%s built an artillery workshop", c.Name)
	} else if c.Buildings[Artillery] > 0 && c.Artillery < 3 && c.Gold >= 8 && c.Wood >= 5 {
		c.Gold -= 8
		c.Wood -= 5
		c.Artillery++
		return fmt.Sprintf("%s trained artillery", c.Name)
	} else if c.Buildings[Forge] == 0 && c.Troops >= 200 && canAfford(c, Forge) {
		g.aiBuild(c, Forge)
		return fmt.Sprintf("%s built a forge", c.Name)
	} else if c.Buildings[Forge] > 0 && c.ForgeLevel < 2 && c.Gold >= 100 && c.Wood >= 40 {
		c.Gold -= 100
		c.Wood -= 40
		c.ForgeLevel++
		return fmt.Sprintf("%s upgraded its army", c.Name)
	} else if c.Buildings[Farm] == 0 && canAfford(c, Farm) {
		g.aiBuild(c, Farm)
		return fmt.Sprintf("%s built a farm", c.Name)
	} else if c.Buildings[Sawmill] == 0 && canAfford(c, Sawmill) {
		g.aiBuild(c, Sawmill)
		return fmt.Sprintf("%s built a sawmill", c.Name)
	} else if c.Buildings[Market] == 0 && canAfford(c, Market) {
		g.aiBuild(c, Market)
		return fmt.Sprintf("%s built a market", c.Name)
	} else if c.Buildings[Barracks] == 0 && canAfford(c, Barracks) {
		g.aiBuild(c, Barracks)
		return fmt.Sprintf("%s built barracks", c.Name)
	} else if c.Gold >= 40 && c.Wood >= 12 {
		choices := []Building{Farm, Sawmill, Market}
		building := choices[g.rng.Intn(len(choices))]
		g.aiBuild(c, building)
		return fmt.Sprintf("%s expanded its %s", c.Name, building)
	}
	return ""
}

func (g *Game) aiBuild(c *Castle, building Building) {
	cost := buildingCosts[building]
	if c.Gold < cost.gold || c.Wood < cost.wood {
		return
	}
	c.Gold -= cost.gold
	c.Wood -= cost.wood
	c.Buildings[building]++
}

func canAfford(c *Castle, building Building) bool {
	cost := buildingCosts[building]
	return c.Gold >= cost.gold && c.Wood >= cost.wood
}

func (g *Game) aiConsiderWar(c *Castle) {
	if g.Relations[c.Owner] != Peace || g.rng.Intn(100) >= 15 {
		return
	}
	for _, neighborName := range c.Neighbors {
		if g.PlayerControls(g.Castles[neighborName]) {
			g.Relations[c.Owner] = War
			g.addEvent(fmt.Sprintf("%s declared war on %s.", c.Owner, g.PlayerName))
			return
		}
	}
}

func joinActivity(activity []string) string {
	return strings.Join(activity, "; ")
}

func (g *Game) tryAIAttack(c *Castle) bool {
	if c.Troops < 20 || g.rng.Intn(4) != 0 {
		return false
	}
	for _, neighborName := range c.Neighbors {
		target := g.Castles[neighborName]
		if !g.PlayerControls(target) {
			continue
		}
		committed := c.Troops / 2
		if committed < 10 {
			committed = c.Troops
		}
		g.resolveBattle(c, target, committed)
		g.CheckOutcome()
		return true
	}
	return false
}
