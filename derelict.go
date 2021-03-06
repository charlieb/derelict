package main

import (
	"container/list"
	"log"
	"math/rand"
	"os"
	"time"
)

var Dlog *log.Logger

const (
	NONE int = -1

	SALVAGE = iota
	REPAIR
	CREATE
	ACTIVATE
)
const (
	FLOOR = iota
	WALL
	CONDUIT
	WALL_CONDUIT
	DOOR
	DOOR_CONDUIT
)

////////////////////// AIR /////////////////////////
type Air struct {
	x, y   int
	air    [][]float64
	buffer [][]float64
}

func (a *Air) Init(x, y int) {
	a.x, a.y = x, y
	a.air = make([][]float64, x, x)
	a.buffer = make([][]float64, x, x)
	for i := 0; i < x; i++ {
		a.air[i] = make([]float64, y, y)
		a.buffer[i] = make([]float64, y, y)
	}
}
func (a *Air) ProcessFlow(cells [][]Cell) {
	const ()
	var total, nairs float64
	for i := 0; i < a.x; i++ {
		for j := 0; j < a.y; j++ {
			if cells[i][j].AirFlows() {
				total = 0
				nairs = 0
				Dlog.Printf("   processFlow cell: (%v, %v)\n", i, j)
				for ii := -1; ii <= 1; ii++ {
					for jj := -1; jj <= 1; jj++ {
						if i+ii >= 0 && i+ii < a.x && j+jj >= 0 && j+jj < a.y {
							if cells[i+ii][j+jj].AirFlows() {
								total += a.air[i+ii][j+jj]
								nairs++
								Dlog.Printf("   processFlow (%v, %v), flows %v / %v\n", i+ii, j+jj, total, nairs)
							}
						}
					}
				}
				if nairs == 0 || total == 0 {
					a.buffer[i][j] = cells[i][j].AirSinkSource(0)
				} else {
					a.buffer[i][j] = cells[i][j].AirSinkSource(total / nairs)
				}
				Dlog.Printf("   processFlow (%v, %v) airs: %v, total: %v\n", i, j, nairs, total)
			}
		}
	}
	tmp := a.air
	a.air = a.buffer
	a.buffer = tmp
	Dlog.Println("<- processFlow")
}

////////////////////// ENERGY ////////////////////////
type Energy struct {
	x, y   int
	energy [][]float64
}

func (e *Energy) Init(x, y int) {
	e.x, e.y = x, y
	e.energy = make([][]float64, x, x)
	for i := 0; i < x; i++ {
		e.energy[i] = make([]float64, y, y)
	}
}
func (e *Energy) ProcessFlow(cells [][]Cell) {
	tbd := list.New()
	for i := 0; i < e.x; i++ {
		for j := 0; j < e.y; j++ {
			e.energy[i][j] = cells[i][j].EnergySinkSource(0)
			if e.energy[i][j] > 0 {
				tbd.PushBack([2]int{i, j})
			}
		}
	}
	member := func(el [2]int, lst *list.List) bool {
		for e := lst.Front(); e != nil; e = e.Next() {
			a := e.Value.([2]int)
			if a[0] == el[0] && a[1] == el[1] {
				return true
			}
		}
		return false
	}

	done := list.New()
	var follow func(int, int)
	follow = func(x, y int) {
		for i := x - 1; i <= x+1; i++ {
			for j := y - 1; j <= y+1; j++ {
				if i >= 0 && i < e.x && j >= 0 && j < e.y {
					if cells[i][j].EnergyFlows() && !member([2]int{i, j}, done) {
						Dlog.Printf("Energy.ProcessFlows: (%v, %v)\n", i, j)
						done.PushBack([2]int{i, j})
						if e.energy[x][y] > e.energy[i][j] {
							e.energy[i][j] = cells[i][j].EnergySinkSource(e.energy[x][y])
							if e.energy[i][j] > 0 {
								follow(i, j)
							}
						}
					}
				}
			}
		}
	}
	for e := tbd.Front(); e != nil; e = e.Next() {
		Dlog.Printf("Energy.ProcessFlows Source: (%v, %v)\n", e.Value.([2]int)[0], e.Value.([2]int)[1])
		done.PushBack(e.Value)
		follow(e.Value.([2]int)[0], e.Value.([2]int)[1])
	}
}

////////////////////// LEVEL /////////////////////////

type Level struct {
	x, y           int
	exit_x, exit_y int

	cells [][]Cell
	air   Air

	energy Energy
}

func (level *Level) Init() {
	level.cells = make([][]Cell, level.x, level.x)
	level.air.Init(level.x, level.y)
	level.energy.Init(level.x, level.y)
	for i := 0; i < level.x; i++ {
		level.cells[i] = make([]Cell, level.y, level.y)
		for j := 0; j < level.y; j++ {
			level.cells[i][j] = new(Vacuum)
		}
	}
}
func (level *Level) Iterate() {
	Dlog.Println("-> Level.Iterate")
	level.air.ProcessFlow(level.cells)
	level.energy.ProcessFlow(level.cells)
	Dlog.Println("<- Level.Iterate")
}

type Drawable interface {
	Character() int32
}

////////////////////// PLAYER /////////////////////////
const (
	noSensor = iota
	pressureSensor
	energySensor
	maxSensor
)

type Player struct {
	x, y   int
	vision int

	energy_left, energy_capcacity float64

	sensor                int
	energy_sensor_range   int
	pressure_sensor_range int

	air_left, air_capacity float64
	dead                   bool
	left_ship              bool
	helmet_on              bool

	copper, steel int
}

func (p *Player) Init() {
	p.x, p.y = 1, 1
	p.vision = 5
	p.left_ship = false

	p.energy_left, p.energy_capcacity = 1.0, 1.0

	p.sensor = noSensor
	p.pressure_sensor_range = 2
	p.energy_sensor_range = 1

	p.air_left, p.air_capacity = 10.0, 10.0
	p.helmet_on = true
	p.dead = false
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
func (p *Player) Iterate(level *Level) {
	if !p.left_ship && (level.exit_x != p.x || level.exit_y != p.y) {
		p.left_ship = true
	}

	const med, low float64 = 6, 3
	if level.air.air[p.x][p.y] < low {
		p.air_left -= 0.1 / (1 + level.air.air[p.x][p.y])
		/*
			} else if level.air.air[p.x][p.y] >= low && level.air.air[p.x][p.y] < med {
				// Do nothing, enough air to maintain
		*/
	} else if level.air.air[p.x][p.y] >= med {
		p.air_left += level.air.air[p.x][p.y] / 50
	}

	// Air limits
	if p.air_left <= 0 {
		p.dead = true
	}
	if p.air_left > p.air_capacity {
		p.air_left = p.air_capacity
	}
}

func (p *Player) Action(level *Level, ui UI, action_id int) (turns int) {
	Dlog.Println("-> Player.Action")
	abort := false
	if action_id == NONE {
		action_id, abort = ui.Menu("Choose an Action:",
			[]string{"Salvage",
				"Repair",
				"Create"})
	}
	Dlog.Println("   Player.Action: ", action_id, abort)
	if abort {
		Dlog.Println("-> Player.Action: false")
		return 0
	}
	x, y, abort := ui.DirectionPrompt()
	if abort {
		return 0
	} else if level.cells[p.x+x][p.y+y] != nil {
		replacement := level.cells[p.x+x][p.y+y]
		switch action_id {
		case ACTIVATE:
			turns = level.cells[p.x+x][p.y+y].Activate(ui)
		case SALVAGE:
			turns, replacement = level.cells[p.x+x][p.y+y].Salvage(ui, p)
		case REPAIR:
			turns, replacement = level.cells[p.x+x][p.y+y].Repair(ui, p)
		case CREATE:
			cell, abort := ui.Menu("Create what?",
				[]string{"Floor", "Wall", "Conduit", "Wall/Conduit", "Door", "Door/Conduit"})
			if abort {
				return 0
			}
			var nc Cell
			switch cell {
			case FLOOR:
				nc = new(Floor)
				turns = nc.Create(ui, p)
			case WALL:
				nc = new(Wall)
				turns = nc.Create(ui, p)
			case CONDUIT:
			case WALL_CONDUIT:
			case DOOR:
			case DOOR_CONDUIT:
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

func buildTestLevel(level *Level) {
	x, y := 10, 2
	// Floor
	for i := 0; i < 31; i++ {
		for j := 0; j < 20; j++ {
			level.cells[x+i][y+j] = new(Floor)
		}
	}
	// Outer walls
	for i := 0; i < 31; i++ {
		level.cells[x+i][y+0] = new(Wall)
	}
	for i := 0; i < 31; i++ {
		level.cells[x+i][y+20] = new(Wall)
	}
	for i := 0; i < 20; i++ {
		level.cells[x+0][y+i] = new(Wall)
	}
	for i := 0; i < 20; i++ {
		level.cells[x+30][y+i] = new(Wall)
	}

	// Inner walls
	for i := 0; i < 15; i++ {
		level.cells[x+8][y+i] = new(Wall)
	}
	for i := 0; i < 18; i++ {
		level.cells[x+i][y+14] = new(Wall)
	}
	for i := 15; i < 20; i++ {
		level.cells[x+12][y+i] = new(Wall)
	}
	for i := 8; i < 18; i++ {
		level.cells[x+i][y+8] = new(Wall)
	}
	for i := 0; i < 20; i++ {
		level.cells[x+18][y+i] = new(Wall)
	}
	for i := 0; i < 20; i++ {
		level.cells[x+21][y+i] = new(Wall)
	}
	for i := 21; i < 30; i++ {
		level.cells[x+i][y+10] = new(Wall)
	}

	// Doors
	level.cells[x+0][y+2] = new(Door)
	level.cells[x+8][y+9] = new(Door)
	level.cells[x+14][y+8] = new(Door)
	level.cells[x+14][y+14] = new(Door)
	level.cells[x+12][y+17] = new(Door)
	level.cells[x+18][y+11] = new(Door)
	level.cells[x+21][y+6] = new(Door)
	level.cells[x+21][y+14] = new(Door)

	// Power Plant
	level.cells[x+26][y+3] = new(PowerPlant)
	level.cells[x+27][y+3] = new(PowerPlant)
	level.cells[x+26][y+4] = new(PowerPlant)
	level.cells[x+27][y+4] = new(PowerPlant)
	/*
		level.cells[24][6].(*PowerPlant).damaged = false
		level.cells[25][6].(*PowerPlant).damaged = false
		level.cells[24][7].(*PowerPlant).damaged = false
		level.cells[25][7].(*PowerPlant).damaged = false
	*/

	// Air Plant
	level.cells[x+26][y+17] = new(AirPlant)
	level.cells[x+27][y+17] = new(AirPlant)
	level.cells[x+26][y+16] = new(AirPlant)
	level.cells[x+27][y+16] = new(AirPlant)

	// Conduits
	level.cells[x+28][y+4] = new(Conduit)
	level.cells[x+28][y+5] = new(Conduit)
	level.cells[x+28][y+6] = new(Conduit)
	level.cells[x+28][y+7] = new(Conduit)
	level.cells[x+28][y+8] = new(Conduit)
	level.cells[x+28][y+9] = new(Conduit)
	level.cells[x+28][y+10] = new(WallConduit)
	level.cells[x+28][y+11] = new(Conduit)
	tmp := new(Conduit)
	tmp.damaged = true
	level.cells[x+28][y+12] = tmp
	level.cells[x+28][y+13] = new(Conduit)
	level.cells[x+28][y+14] = new(Conduit)
	level.cells[x+28][y+15] = new(Conduit)
	level.cells[x+28][y+16] = new(Conduit)

	level.cells[0][5] = new(Wall)
	level.cells[1][5] = new(Wall)
	level.cells[0][7] = new(Wall)
	level.cells[1][7] = new(Wall)
	level.cells[1][6] = new(Door)
	level.cells[0][6] = new(EntranceExit)
	level.exit_x, level.exit_y = 0, 6

}

/////////////////// GAME MAIN ///////////////////
type Game struct {
	level  Level
	player Player
	ui     UI
}

func NewGame() Game {
	var game Game
	game.level.x, game.level.y = 69, 23
	game.level.Init()

	buildTestLevel(&game.level)

	game.player.Init()
	game.player.x = 0
	game.player.y = 6

	return game
}
func main() {
	// set a new random seed
	rand.Seed(time.Now().UTC().UnixNano())

	/*
		testLevel()
		return
	*/

	file, err := os.Create("log")
	if err != nil {
		log.Fatal(err)
	}
	Dlog = log.New(file, "DERELICT: ", 0)

	game := NewGame()
	game.ui = NewCursesUI(&game.level, &game.player)
	game.ui.Run()
}
