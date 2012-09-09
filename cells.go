package main

import (
	"fmt"
	"math/rand"
)

////////////////////// CELLS /////////////////////////
type Cell interface {
	Walkable() bool
	SeePast() bool

	// Returns are turns, replacement Cell
	Salvage(UI, *Player) (int, Cell)
	Repair(UI, *Player) (int, Cell)
	Create(UI, *Player) int
}

type Vacuum struct{}

func (c *Vacuum) Walkable() bool   { return true }
func (c *Vacuum) SeePast() bool    { return true }
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

type Floor struct{}

func (c *Floor) Walkable() bool   { return true }
func (c *Floor) SeePast() bool    { return true }
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

type Wall struct {
	damaged bool
}

func (w *Wall) Walkable() bool   { return false }
func (w *Wall) Character() int32 { return '#' }
func (w *Wall) SeePast() bool    { return false }
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
