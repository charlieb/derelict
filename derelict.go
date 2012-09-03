package main

import (
	"container/list"
	"log"
	"os"
)

var Dlog *log.Logger

////////////////////// LEVEL /////////////////////////

type Level struct {
	x, y      int
	items     [][]*list.List
	cells     [][]Cell
	air       [][]float64
	airBuffer [][]float64
}

func (level *Level) Init() {
	level.items = make([][]*list.List, level.x, level.x)
	level.cells = make([][]Cell, level.x, level.x)
	level.air = make([][]float64, level.x, level.x)
	level.airBuffer = make([][]float64, level.x, level.x)
	for i := 0; i < level.x; i++ {
		level.cells[i] = make([]Cell, level.y, level.y)
		level.items[i] = make([]*list.List, level.y, level.y)
		level.air[i] = make([]float64, level.y, level.y)
		level.airBuffer[i] = make([]float64, level.y, level.y)
		for j := 0; j < level.y; j++ {
			level.items[i][j] = list.New()
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

////////////////////// SENSORS /////////////////////////

type Sensor interface {
	Sense(level *Level, player *Player) float64
}

type PressureSensor struct{}

func (s *PressureSensor) Name() string        { return "Pressure Sensor" }
func (s *PressureSensor) Description() string { return "Short range ambient air pressure sensor" }
func (s *PressureSensor) Activate(level *Level, player *Player) {
	// Add pressure sensor to the suit
}

////////////////////// SUIT /////////////////////////

type Suit struct {
	air           float64
	airCapacity   float64
	faceplateOpen bool
}

////////////////////// PLAYER /////////////////////////

type Player struct {
	x, y  int
	items []Item

	vision int
}

func (p *Player) Move(to_x, to_y int) {
	//	for e := level.items[p.pos[0]][p.pos[1]].Front(); e != nil; e = e.Next() {
	//		if e.Value == p {
	//			level.items[p.pos[0]][p.pos[1]].Remove(e)
	//			break
	//		}
	//	}
	Dlog.Println("-> Move", to_x, to_y)
	p.x, p.y = to_x, to_y
	Dlog.Println("<- Move", p.x, p.y)
	//	level.items[p.pos[0]][p.pos[1]].PushFront(p)
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

	game.player.x, game.player.y, game.player.vision = 1, 1, 5

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
