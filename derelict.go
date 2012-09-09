package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
)

var Dlog *log.Logger

////////////////////// LEVEL /////////////////////////

type Level struct {
	x, y      int
	cells     [][]Cell
	air       [][]float64
	airBuffer [][]float64
}

func (level *Level) Init() {
	level.cells = make([][]Cell, level.x, level.x)
	level.air = make([][]float64, level.x, level.x)
	level.airBuffer = make([][]float64, level.x, level.x)
	for i := 0; i < level.x; i++ {
		level.cells[i] = make([]Cell, level.y, level.y)
		level.air[i] = make([]float64, level.y, level.y)
		level.airBuffer[i] = make([]float64, level.y, level.y)
		for j := 0; j < level.y; j++ {
			level.cells[i][j] = new(Floor)
		}
	}
}
func (level *Level) outerWall() {
	for i := 0; i < level.x; i++ {
		level.cells[i][0] = new(Wall)
	}
	for i := 0; i < level.x; i++ {
		level.cells[i][level.y-1] = new(Wall)
	}
	for j := 0; j < level.y; j++ {
		level.cells[0][j] = new(Wall)
	}
	for j := 0; j < level.y; j++ {
		level.cells[level.x-1][j] = new(Wall)
	}
}
func (level *Level) Iterate(its int) {
	Dlog.Println("-> Level.Iterate")
	for it := 0; it < its; it++ {
		for i := 0; i < level.x; i++ {
			for j := 0; j < level.y; j++ {
				level.airBuffer[i][j] = 0
			}
		}

		var total, nairs float64
		for i := 0; i < level.x; i++ {
			for j := 0; j < level.y; j++ {
				total = 0
				nairs = 0
				Dlog.Printf("   Level.Iterate cell: (%v, %v)\n", i, j)
				for ii := -1; ii <= 1; ii++ {
					for jj := -1; jj <= 1; jj++ {
						if i+ii >= 0 && i+ii < level.x && j+jj >= 0 && j+jj < level.y {
							Dlog.Printf("   Level.Iterate (%v, %v)\n", i+ii, j+jj)
							if level.cells[i+ii][j+jj] == nil || level.cells[i+ii][j+jj].Walkable() {
								total += level.air[i+ii][j+jj]
								nairs++
							}
						}
					}
				}
				if nairs == 0 || total == 0 {
					continue
				}
				Dlog.Printf("   Level.Iterate (%v, %v) airs: %v, total: %v\n", i, j, nairs, total)
				level.airBuffer[i][j] = total / nairs
			}
		}
		tmp := level.air
		level.air = level.airBuffer
		level.airBuffer = tmp
	}
	Dlog.Println("<- Level.Iterate")
}

type Drawable interface {
	Character() int32
}

////////////////////// PLAYER /////////////////////////

type Player struct {
	x, y   int
	vision int

	energy_left, energy_capcacity float64

	pressure_sensor_range int
	pressure_sensor_on    bool

	air_left, air_capacity float64
	helmet_on              bool

	copper, steel int
}

func (p *Player) Init() {
	p.x, p.y = 1, 1
	p.vision = 5

	p.energy_left, p.energy_capcacity = 1.0, 1.0

	p.pressure_sensor_range = 1
	p.pressure_sensor_on = false

	p.air_left, p.air_capacity = 1.0, 1.0
	p.helmet_on = true
}
func (p *Player) Move(to_x, to_y int) {
	Dlog.Println("-> Move", to_x, to_y)
	p.x, p.y = to_x, to_y
	Dlog.Println("<- Move", p.x, p.y)
}
func (p *Player) Walk(dir_x, dir_y int, level *Level) bool {
	px, py := p.x+dir_x, p.y+dir_y
	Dlog.Println("-> Walk", px, py)
	if px >= 0 && px < level.x && py >= 0 && py < level.y {
		if level.cells[px][py] == nil || level.cells[px][py].Walkable() {
			p.Move(px, py)
			Dlog.Println("<- Walk", true)
			return true
		}
	}
	Dlog.Println("<- Walk", false)
	return false
}
func (p *Player) Character() int32 { return '@' }

func (p *Player) buildWall(x, y int) {

}
func (p *Player) Action(level *Level, ui UI) int { // number of turns
	Dlog.Println("-> Player.Action")
	action, abort := ui.Menu("Choose an Action:",
		[]string{"Salvage",
			"Repair",
			"Create"})
	var (
		turns       int
		replacement Cell
	)
	Dlog.Println("   Player.Action: ", action, abort)
	if abort {
		Dlog.Println("-> Player.Action: false")
		return 0
	}
	x, y, abort := ui.DirectionPrompt()
	if abort {
		return 0
	} else if level.cells[p.x+x][p.y+y] != nil {
		switch action {
		case 0: // Salvage
			turns, replacement = level.cells[p.x+x][p.y+y].Salvage(ui, p)
		case 1: // Repair
			turns, replacement = level.cells[p.x+x][p.y+y].Repair(ui, p)
		case 2: // Create
			action, abort := ui.Menu("Create what?", []string{"Floor", "Wall", "Conduit", "Wall/Conduit", "Door"})
			if abort {
				return 0
			}
			var nc Cell
			switch action {
			case 0: // Floor
				nc = new(Floor)
				turns = nc.Create(ui, p)
			case 1: // Wall
				nc = new(Wall)
				turns = nc.Create(ui, p)
			case 2: // Conduit
			case 3: // Wall/Conduit
			case 4: // Door
			}
			if turns > 0 {
				replacement = nc
			}
		}
		level.cells[p.x+x][p.y+y] = replacement
	}

	Dlog.Println("<- Player.Action: true")
	return turns
}

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
func buildTestWalls(level *Level) {
	for i := 2; i < 5; i++ {
		level.cells[i][2] = new(Wall)
	}
	for j := 3; j < 5; j++ {
		level.cells[2][j] = new(Wall)
	}
}

/////////////////// ///////////////////
type Game struct {
	level  Level
	player Player
	ui     UI
}

func NewGame() Game {
	var game Game
	game.level.x, game.level.y = 69, 23
	game.level.Init()
	game.level.outerWall()
	game.level.air[10][10] = 9
	game.level.air[10][11] = 9

	buildTestWalls(&game.level)

	game.player.Init()

	return game
}
func main() {
	file, err := os.Create("log")
	if err != nil {
		log.Fatal(err)
	}
	Dlog = log.New(file, "DERELICT: ", 0)

	game := NewGame()
	game.ui = new(CursesUI)
	game.ui.Run(&game)
}
