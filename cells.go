package main

import (
	"fmt"
	"math/rand"
)

////////////////////// CELLS /////////////////////////
type Cell interface {
	Description() string
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
			ui.Message(fmt.Sprintf("Used %v steel and %v copper to repair the %v in %v turns",
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

func (c *Vacuum) Description() string                { return "The cold vacuum of space" }
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

func (c *Floor) Description() string                { return "The floor" }
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

func (c *Wall) Description() string                { return "A wall" }
func (w *Wall) Walkable() bool                     { return false }
func (w *Wall) Character() int32                   { return '#' }
func (w *Wall) SeePast() bool                      { return false }
func (w *Wall) AirFlows() bool                     { return w.damaged }
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

func (c *Door) Description() string { return "A door" }
func (d *Door) Walkable() bool      { return d.open }
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
	damaged bool
}

func (c *Conduit) Description() string {
	if c.damaged {
		return "A burned out energy conduit"
	}
	return "An energy conduit"
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

//////////////// WALL CONDUIT /////////////////////

type WallConduit struct {
	damaged bool
}

func (c *WallConduit) Description() string {
	if c.damaged {
		return "A burned out energy conduit passes through a wall here"
	}
	return "An energy conduit passes through a wall here"
}
func (c *WallConduit) Walkable() bool                     { return false }
func (c *WallConduit) SeePast() bool                      { return false }
func (c *WallConduit) AirFlows() bool                     { return c.damaged }
func (c *WallConduit) AirSinkSource(a float64) float64    { return a }
func (c *WallConduit) EnergyFlows() bool                  { return !c.damaged }
func (c *WallConduit) EnergySinkSource(e float64) float64 { return e }
func (c *WallConduit) Character() int32 {
	if c.damaged {
		return '%'
	}
	return '*'
}
func (c *WallConduit) Salvage(ui UI, p *Player) (int, Cell) {
	return genericSalvage(10, 10, 15, ui, p), new(Floor)
}
func (c *WallConduit) Repair(ui UI, p *Player) (int, Cell) {
	return genericRepair(&c.damaged, 10, 10, 15, "conduit", ui, p), c
}
func (c *WallConduit) Create(ui UI, p *Player) int {
	return genericCreate(15, 15, 15, "conduit", ui, p)
}
func (c *WallConduit) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 1
}

///////////// POWER PLANT /////////////////

type PowerPlant struct {
	damaged bool
}

func (c *PowerPlant) Description() string             { return "An energy generator" }
func (c *PowerPlant) Walkable() bool                  { return false }
func (c *PowerPlant) SeePast() bool                   { return false }
func (c *PowerPlant) AirFlows() bool                  { return false }
func (c *PowerPlant) AirSinkSource(a float64) float64 { return a }
func (c *PowerPlant) EnergyFlows() bool               { return !c.damaged }
func (c *PowerPlant) EnergySinkSource(e float64) float64 {
	if !c.damaged {
		return 9
	}
	return e
}
func (c *PowerPlant) Character() int32 {
	if c.damaged {
		return 'p'
	}
	return 'P'
}
func (c *PowerPlant) Salvage(ui UI, p *Player) (int, Cell) {
	return genericSalvage(10, 10, 20, ui, p), new(Floor)
}
func (c *PowerPlant) Repair(ui UI, p *Player) (int, Cell) {
	return genericRepair(&c.damaged, 10, 10, 15, "power plant", ui, p), c
}
func (c *PowerPlant) Create(ui UI, p *Player) int {
	ui.Message("You cannot create a power plant from scratch")
	return 0
}
func (c *PowerPlant) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 1
}

///////////// AIR PLANT /////////////////

type AirPlant struct {
	damaged bool
	energy  float64
}

func (c *AirPlant) Description() string { return "An air generator" }
func (c *AirPlant) Walkable() bool      { return false }
func (c *AirPlant) SeePast() bool       { return false }
func (c *AirPlant) AirFlows() bool      { return !c.damaged }
func (c *AirPlant) AirSinkSource(a float64) float64 {
	Dlog.Println("<> AirPlant")
	if !c.damaged && c.energy > 5 {
		return 9
	}
	return a
}
func (c *AirPlant) EnergyFlows() bool { return true }
func (c *AirPlant) EnergySinkSource(e float64) float64 {
	c.energy = e
	if e > 5 {
		return e - (e - 5) / 50
	}
	return e
}
func (c *AirPlant) Character() int32 {
	if c.damaged {
		return 'a'
	}
	return 'A'
}
func (c *AirPlant) Salvage(ui UI, p *Player) (int, Cell) {
	return genericSalvage(10, 10, 20, ui, p), new(Floor)
}
func (c *AirPlant) Repair(ui UI, p *Player) (int, Cell) {
	return genericRepair(&c.damaged, 10, 10, 15, "air plant", ui, p), c
}
func (c *AirPlant) Create(ui UI, p *Player) int {
	ui.Message("You cannot create a air plant from scratch")
	return 0
}
func (c *AirPlant) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 1
}
///////////// Entrance/Exit /////////////////

type EntranceExit struct {
}

func (c *EntranceExit) Description() string { return "Your ship, safety" }
func (c *EntranceExit) Walkable() bool      { return true }
func (c *EntranceExit) SeePast() bool       { return true }
func (c *EntranceExit) AirFlows() bool      { return true }
func (c *EntranceExit) AirSinkSource(a float64) float64 { return 9 }
func (c *EntranceExit) EnergyFlows() bool { return false }
func (c *EntranceExit) EnergySinkSource(e float64) float64 { return 0 }
func (c *EntranceExit) Character() int32 { return '.' }
func (c *EntranceExit) Salvage(ui UI, p *Player) (int, Cell) {
	ui.Message("Why would you salvage your own ship?")
	return 0, c
}
func (c *EntranceExit) Repair(ui UI, p *Player) (int, Cell) {
	ui.Message("Your own ship does not need repair")
	return 0, c
}
func (c *EntranceExit) Create(ui UI, p *Player) int {
	Dlog.Println("<> BUG EntranceExit.Create")
	ui.Message("Create shold never be called on an EntranceExit cell")
	return 0
}
func (c *EntranceExit) Activate(ui UI) int {
	ui.Message("Nothing happens")
	return 1
}
