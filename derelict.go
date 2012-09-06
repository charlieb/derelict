package main

import (
	"log"
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
func (level *Level) Iterate() {
	Dlog.Println("-> Level.Iterate")
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
	Dlog.Println("<- Level.Iterate")
}

type Cell interface {
	Walkable() bool

	SeePast() bool
}
type Item interface {
	Name() string
	Description() string
	Activate(*Level, *Player)
}
type Drawable interface {
	Character() int32
}

////////////////////// PLAYER /////////////////////////

type Player struct {
	x, y  int
	vision int

	energy_left, energy_capcacity float64

	pressure_sensor_range int
	pressure_sensor_on bool

	air_left, air_capacity float64
	helmet_on bool
}
func (p *Player) Init() {
	p.x, p.y = 1,1
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
func (p *Player) Action(level *Level, ui UI) bool {
	Dlog.Println("-> Player.Action")
	action, abort := ui.Menu([]string{"Activate",
		"-",
		"Build Wall",
		"Repair Wall"})
		Dlog.Println("   Player.Action: ", action, abort)
	if abort {
		Dlog.Println("-> Player.Action: false")
		return false
	}
	switch action {
	case 0: // Activate
		x, y, abort := ui.DirectionPrompt()
		if !abort && level.cells[p.x + x][p.y + y] == nil {
		}
	case 1: // Build Wall
		x, y, abort := ui.DirectionPrompt()
		if abort {
			return false
		} else if level.cells[p.x+x][p.y+y] == nil {
			level.cells[p.x+x][p.y+y] = new(Wall)
			ui.Message("Built a wall")
		} else {
			ui.Message("That square is occupied")
		}
	case 2: // Repair Wall
		ui.Message("Repair a wall!")
	}
	Dlog.Println("<- Player.Action: true")
	return true
}

////////////////////// /////////////////////////

type sCell struct {
	visible bool
}

func (c *sCell) Walkable() bool { return true }
func (c *sCell) SeePast() bool  { return true }

type Wall struct {
	sCell
}

func (w *Wall) Walkable() bool   { return false }
func (w *Wall) Character() int32 { return '#' }
func (w *Wall) SeePast() bool    { return false }

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
