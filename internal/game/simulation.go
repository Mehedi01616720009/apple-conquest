package game

import "fmt"

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
			g.aiDevelop(c)
			g.aiConsiderWar(c)
		}
	}
	if g.Tick%10 == 0 {
		g.addEvent(fmt.Sprintf("Day %d: the realm continues to develop.", g.Tick/10))
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

func (g *Game) aiDevelop(c *Castle) {
	if g.Relations[c.Owner] == War && g.tryAIAttack(c) {
		return
	}
	if c.Buildings[Barracks] > 0 && c.Troops < 180 && c.Gold >= 30 && c.Wood >= 15 && c.Food >= 10 {
		c.Gold -= 30
		c.Wood -= 15
		c.Food -= 10
		c.Troops += 10
	} else if c.Buildings[Artillery] == 0 && c.Troops >= 180 && canAfford(c, Artillery) {
		g.aiBuild(c, Artillery)
	} else if c.Buildings[Artillery] > 0 && c.Artillery < 3 && c.Gold >= 8 && c.Wood >= 5 {
		c.Gold -= 8
		c.Wood -= 5
		c.Artillery++
	} else if c.Buildings[Forge] == 0 && c.Troops >= 200 && canAfford(c, Forge) {
		g.aiBuild(c, Forge)
	} else if c.Buildings[Forge] > 0 && c.ForgeLevel < 2 && c.Gold >= 100 && c.Wood >= 40 {
		c.Gold -= 100
		c.Wood -= 40
		c.ForgeLevel++
	} else if c.Buildings[Farm] == 0 && canAfford(c, Farm) {
		g.aiBuild(c, Farm)
	} else if c.Buildings[Sawmill] == 0 && canAfford(c, Sawmill) {
		g.aiBuild(c, Sawmill)
	} else if c.Buildings[Market] == 0 && canAfford(c, Market) {
		g.aiBuild(c, Market)
	} else if c.Buildings[Barracks] == 0 && canAfford(c, Barracks) {
		g.aiBuild(c, Barracks)
	} else if c.Gold >= 40 && c.Wood >= 12 {
		choices := []Building{Farm, Sawmill, Market}
		g.aiBuild(c, choices[g.rng.Intn(len(choices))])
	}
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
	if g.Relations[c.Owner] != Peace || g.rng.Intn(100) != 0 {
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
