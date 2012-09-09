package main

import (
	"fmt"
	"math/rand"
)

////////////////////// CELLS /////////////////////////
type Cell interface {
	Walkable() bool
	SeePast() bool
	AirFlows() bool

	// Returns are turns, replacement Cell
	Salvage(UI, *Player) (int, Cell)
	Repair(UI, *Player) (int, Cell)
	Create(UI, *Player) int

	Activate(UI) int

	Iterate(int, int, *Level) // x, y positions and the level
}

type Vacuum struct{}

func (c *Vacuum) Walkable() bool   { return true }
func (c *Vacuum) SeePast() bool    { return true }
func (c *Vacuum) AirFlows() bool    { return true }
func (c *Vacuum) Character() int32 { return ' ' }
func (c *Vacuum) Salvage(ui UI, p *Player) (int, Cell) {
	ui.Message("There is nothing to salvage in a vacuum")
	return 0, c
}
func (c *Vacuum) Repair(ui UI, p *Player) (int, Cell) {
	ui.Message("You cannot repair a vacuum")
	return 0, c
}
func (c *Vacuum) Create(ui UI, p *Player) int {
	ui.Message("Nature abhors a vacuum")
	return 0
}
func (c *Vacuum) Activate(ui UI) int {
	ui.Message("You activate the vacuum, the universe is re-created in a flash")
	return 1
}
func (c *Vacuum) Iterate(x, y int, level *Level) { level.air[x][y] = 0 }

///////////////// FLOOR /////////////////
type Floor struct{}

func (c *Floor) Walkable() bool   { return true }
func (c *Floor) SeePast() bool    { return true }
func (c *Floor) AirFlows() bool    { return true }
func (c *Floor) Character() int32 { return '.' }
func (c *Floor) Salvage(ui UI, p *Player) (turns int, replacement Cell) {
	turns = 0
	replacement = c

	sure, aborted := ui.YesNoPrompt("Salvage floor?")
	if !aborted && sure {
		st := rand.Intn(10)
		p.steel += st
		turns = 1 + rand.Intn(9)
		replacement = new(Vacuum)
		ui.Message(fmt.Sprintf("You salvage %v steel in %v turns", st, turns))
	}
	return
}
func (c *Floor) Repair(ui UI, p *Player) (int, Cell) {
	ui.Message("The floor does not need to be repaired")
	return 0, c
}
func genericCreate(max_steel, max_copper, max_turns int, name string, ui UI, p *Player) (turns int) {
	var st, cu int = 0, 0
	if max_steel > 0 {
		st = rand.Intn(max_steel)
	}
	if max_copper > 0 {
		cu = rand.Intn(max_copper)
	}
	turns = 1 + rand.Intn(max_turns-1)

	p.steel -= st
	p.copper -= cu
	if p.steel < 0 && p.copper < 0 {
		ui.Message(fmt.Sprintf("You run out of both steel and copper after %v turns", turns))
		p.steel = 0
		p.copper = 0
	} else if p.steel < 0 {
		ui.Message(fmt.Sprintf("You run out of steel after %v turns", turns))
		p.steel = 0
	} else if p.copper < 0 {
		ui.Message(fmt.Sprintf("You run out of copper after %v turns", turns))
		p.copper = 0
	} else {
		ui.Message(fmt.Sprintf("Used %v steel and %v copper to create a %v section in %v turns",
			st, cu, name, turns))
	}
	return
}
func (c *Floor) Create(ui UI, p *Player) int {
	return genericCreate(10, 0, 10, "floor", ui, p)
}
func (c *Floor) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 0
}
func (c *Floor) Iterate(x, y int, l *Level) { }

//////////// WALL //////////////////
type Wall struct {
	damaged bool
}

func (w *Wall) Walkable() bool   { return false }
func (w *Wall) Character() int32 { return '#' }
func (w *Wall) SeePast() bool    { return false }
func (w *Wall) AirFlows() bool    { return !w.damaged }
func (c *Wall) Salvage(ui UI, p *Player) (turns int, replacement Cell) {
	turns = 0
	replacement = c

	st := rand.Intn(10)
	p.steel += st
	turns = 1 + rand.Intn(9)
	replacement = new(Floor)
	ui.Message(fmt.Sprintf("You salvage %v steel in %v turns", st, turns))
	return
}
func (c *Wall) Repair(ui UI, p *Player) (turns int, replacement Cell) {
	turns = 1 // Inpecting the wall takes at least 1 turn
	replacement = c

	if c.damaged {
		st := rand.Intn(5)
		p.steel -= st
		turns = 1 + rand.Intn(4)
		if p.steel < 0 {
			ui.Message(fmt.Sprintf("You run out of steel after %v turns", turns))
			p.steel = 0
		} else {
			c.damaged = false
			ui.Message(fmt.Sprintf("Used %v steel to repair the wall in %v turns", st, turns))
		}
	} else {
		ui.Message("The wall does not need to be repaired")
	}
	return
}
func (c *Wall) Create(ui UI, p *Player) (turns int) {
	return genericCreate(10, 0, 10, "wall", ui, p)
}
func (c *Wall) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 0
}
func (c *Wall) Iterate(x, y int, l *Level) { }

/////////// DOOR ////////////////////
type Door struct {
	open, damaged bool
}

func (d *Door) Walkable() bool { return d.open }
func (d *Door) Character() int32 {
	if d.open {
		return '/'
	}
	return '+'
}
func (d *Door) SeePast() bool { return d.open }
func (d *Door) AirFlows() bool { return d.open || d.damaged }
func (c *Door) Salvage(ui UI, p *Player) (turns int, replacement Cell) {
	turns = 0
	replacement = c

	st := rand.Intn(10)
	cu := rand.Intn(10)
	p.steel += st
	p.copper += cu
	turns = 1 + rand.Intn(14)
	replacement = new(Floor)
	ui.Message(fmt.Sprintf("You salvage %v steel and %v copper in %v turns", st, cu, turns))
	return
}
func (c *Door) Repair(ui UI, p *Player) (turns int, replacement Cell) {
	turns = 1 // Inpecting the door takes at least 1 turn
	replacement = c

	if c.damaged {
		st := rand.Intn(5)
		cu := rand.Intn(5)
		p.steel -= st
		p.copper -= cu
		turns = 1 + rand.Intn(9)
		if p.steel < 0 && p.copper < 0 {
			ui.Message(fmt.Sprintf("You run out of steel and copper after %v turns", turns))
			p.steel = 0
			p.copper = 0
		} else if p.steel < 0 {
			ui.Message(fmt.Sprintf("You run out of steel after %v turns", turns))
			p.steel = 0
		} else if p.copper < 0 {
			ui.Message(fmt.Sprintf("You run out of copper after %v turns", turns))
			p.copper = 0
		} else {
			c.damaged = false
			ui.Message(fmt.Sprintf("Used %v steel and $v copper to repair the door in %v turns",
				st, cu, turns))
		}
	} else {
		ui.Message("The door does not need to be repaired")
	}
	return
}
func (c *Door) Create(ui UI, p *Player) (turns int) {
	return genericCreate(10, 0, 10, "wall", ui, p)
}
func (c *Door) Activate(ui UI) int {
	if c.damaged {
		ui.Message("The door is damaged and will not move")
		return 1
	}
	if c.open {
		ui.Message("The door closes")
	} else {
		ui.Message("The door opens")
	}
	c.open = !c.open
	return 1
}
func (c *Door) Iterate(x, y int, l *Level) { }
