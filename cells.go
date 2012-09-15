package main

import (
	"fmt"
	"math/rand"
)

////////////////////// CELLS /////////////////////////
type Cell interface {
	Walkable() bool
	SeePast() bool

	// Air
	AirFlows() bool
	AirSinkSource(float64) float64 // Each cell can adjust its amount of air
	// Energy
	EnergyFlows() bool
	EnergySinkSource(float64) float64 // Each cell can adjust its amount of energy

	// Returns are turns, replacement Cell
	Salvage(UI, *Player) (int, Cell)
	Repair(UI, *Player) (int, Cell)
	Create(UI, *Player) int

	Activate(UI) int
}

////////////// GENERIC ////////////////////
func genericSalvage(max_steel, max_copper, max_turns int, ui UI, p *Player) (turns int) {
	var st, cu int = 0, 0
	if max_steel > 0 {
		st = rand.Intn(max_steel)
	}
	if max_copper > 0 {
		cu = rand.Intn(max_copper)
	}
	turns = 1 + rand.Intn(max_turns-1)

	p.steel += st
	p.copper += cu
	if st == 0 && cu == 0 {
		ui.Message(fmt.Sprintf("You fail to salvage any useful metals", turns))
	} else if cu == 0 {
		ui.Message(fmt.Sprintf("You salvage %v steel in %v turns", st, turns))
	} else if st == 0 {
		ui.Message(fmt.Sprintf("You salvage %v copper in %v turns", cu, turns))
	} else {
		ui.Message(fmt.Sprintf("You salvage %v steel and %v copper in %v turns", st, cu, turns))
	}
	return
}
func genericRepair(damaged *bool, max_steel, max_copper, max_turns int, name string, ui UI, p *Player) (turns int) {
	turns = 1 // Inpecting the "name" takes at least 1 turn

	if *damaged {
		var st, cu int = 0, 0
		if max_steel > 0 {
			st = rand.Intn(max_steel)
		}
		if max_copper > 0 {
			cu = rand.Intn(max_copper)
		}
		p.steel -= st
		p.copper -= cu
		turns = 1 + rand.Intn(max_turns-1)
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
			*damaged = false
			ui.Message(fmt.Sprintf("Used %v steel and $v copper to repair the %v in %v turns",
				st, cu, name, turns))
		}
	} else {
		ui.Message(fmt.Sprintf("The %v does not need to be repaired", name))
	}
	return
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

///////////// VACUUM ////////////////////
type Vacuum struct{}

func (c *Vacuum) Walkable() bool                     { return true }
func (c *Vacuum) SeePast() bool                      { return true }
func (c *Vacuum) AirFlows() bool                     { return true }
func (c *Vacuum) AirSinkSource(float64) float64      { return 0 }
func (c *Vacuum) EnergyFlows() bool                  { return false }
func (c *Vacuum) EnergySinkSource(e float64) float64 { return e }
func (c *Vacuum) Character() int32                   { return ' ' }
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

///////////////// FLOOR /////////////////
type Floor struct{}

func (c *Floor) Walkable() bool                     { return true }
func (c *Floor) SeePast() bool                      { return true }
func (c *Floor) AirFlows() bool                     { return true }
func (c *Floor) AirSinkSource(a float64) float64    { return a }
func (c *Floor) EnergyFlows() bool                  { return false }
func (c *Floor) EnergySinkSource(e float64) float64 { return e }
func (c *Floor) Character() int32                   { return '.' }
func (c *Floor) Salvage(ui UI, p *Player) (turns int, replacement Cell) {
	turns = 0
	replacement = c

	sure, aborted := ui.YesNoPrompt("Salvage floor?")
	if !aborted && sure {
		turns = genericSalvage(10, 0, 10, ui, p)
		replacement = new(Vacuum)
	}
	return
}
func (c *Floor) Repair(ui UI, p *Player) (int, Cell) {
	ui.Message("The floor does not need to be repaired")
	return 0, c
}

func (c *Floor) Create(ui UI, p *Player) int {
	return genericCreate(10, 0, 10, "floor", ui, p)
}
func (c *Floor) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 0
}

//////////// WALL //////////////////
type Wall struct {
	damaged bool
}

func (w *Wall) Walkable() bool                     { return false }
func (w *Wall) Character() int32                   { return '#' }
func (w *Wall) SeePast() bool                      { return false }
func (w *Wall) AirFlows() bool                     { return !w.damaged }
func (c *Wall) AirSinkSource(a float64) float64    { return a }
func (w *Wall) EnergyFlows() bool                  { return false }
func (c *Wall) EnergySinkSource(e float64) float64 { return e }
func (c *Wall) Salvage(ui UI, p *Player) (turns int, replacement Cell) {
	turns = genericSalvage(10, 0, 10, ui, p)
	replacement = new(Floor)
	return
}
func (c *Wall) Repair(ui UI, p *Player) (turns int, replacement Cell) {
	return genericRepair(&c.damaged, 5, 0, 5, "wall", ui, p), c
}
func (c *Wall) Create(ui UI, p *Player) (turns int) {
	return genericCreate(10, 0, 10, "wall", ui, p)
}
func (c *Wall) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 0
}

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
func (d *Door) SeePast() bool                      { return d.open }
func (d *Door) AirFlows() bool                     { return d.open || d.damaged }
func (c *Door) AirSinkSource(a float64) float64    { return a }
func (d *Door) EnergyFlows() bool                  { return false }
func (c *Door) EnergySinkSource(e float64) float64 { return e }
func (c *Door) Salvage(ui UI, p *Player) (int, Cell) {
	return genericSalvage(10, 10, 15, ui, p), new(Floor)
}
func (c *Door) Repair(ui UI, p *Player) (int, Cell) {
	return genericRepair(&c.damaged, 5, 5, 10, "door", ui, p), c
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

//////////////// CONDUIT /////////////////////

type Conduit struct {
	energy  float32
	damaged bool
}

func (c *Conduit) Walkable() bool                     { return true }
func (c *Conduit) SeePast() bool                      { return true }
func (c *Conduit) AirFlows() bool                     { return true }
func (c *Conduit) AirSinkSource(a float64) float64    { return a }
func (c *Conduit) EnergyFlows() bool                  { return !c.damaged }
func (c *Conduit) EnergySinkSource(e float64) float64 { return e }
func (c *Conduit) Character() int32 {
	if c.damaged {
		return '~'
	}
		return '-'
}
func (c *Conduit) Salvage(ui UI, p *Player) (int, Cell) {
	return genericSalvage(0, 10, 10, ui, p), new(Floor)
}
func (c *Conduit) Repair(ui UI, p *Player) (int, Cell) {
	return genericRepair(&c.damaged, 0, 10, 5, "conduit", ui, p), c
}
func (c *Conduit) Create(ui UI, p *Player) int {
	return genericCreate(0, 15, 10, "conduit", ui, p)
}
func (c *Conduit) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 1
}
