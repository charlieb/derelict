package main

import (
	"log"
	"os"
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

////////////////////// LEVEL /////////////////////////

type Level struct {
	x, y      int
	cells     [][]Cell
	air       [][]float64
	airBuffer [][]float64

	energy       [][]float64
	energyBuffer [][]float64
}

func (level *Level) Init() {
	level.cells = make([][]Cell, level.x, level.x)
	level.air = make([][]float64, level.x, level.x)
	level.airBuffer = make([][]float64, level.x, level.x)
	level.energy = make([][]float64, level.x, level.x)
	level.energyBuffer = make([][]float64, level.x, level.x)
	for i := 0; i < level.x; i++ {
		level.cells[i] = make([]Cell, level.y, level.y)
		level.air[i] = make([]float64, level.y, level.y)
		level.airBuffer[i] = make([]float64, level.y, level.y)
		level.energy[i] = make([]float64, level.y, level.y)
		level.energyBuffer[i] = make([]float64, level.y, level.y)
		for j := 0; j < level.y; j++ {
			level.cells[i][j] = new(Vacuum)
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
func (level *Level) processFlow(flow, flowBuffer *[][]float64,
	flowsp func(Cell) bool,
	sinksource func(Cell, float64) float64) {
		Dlog.Println("-> processFlow")
	const min_flow, max_flow float64 = 0, 9
	var total, nairs float64
	for i := 0; i < level.x; i++ {
		for j := 0; j < level.y; j++ {
			if flowsp(level.cells[i][j]) {
				total = 0
				nairs = 0
				Dlog.Printf("   processFlow cell: (%v, %v)\n", i, j)
				for ii := -1; ii <= 1; ii++ {
					for jj := -1; jj <= 1; jj++ {
						if i+ii >= 0 && i+ii < level.x && j+jj >= 0 && j+jj < level.y {
							if flowsp(level.cells[i+ii][j+jj]) {
								total += (*flow)[i+ii][j+jj]
								nairs++
								Dlog.Printf("   processFlow (%v, %v), flows %v / %v\n", i+ii, j+jj, total, nairs)
							}
						}
					}
				}
				if nairs == 0 || total == 0 {
					(*flowBuffer)[i][j] = sinksource(level.cells[i][j], 0)
				} else {
					(*flowBuffer)[i][j] = sinksource(level.cells[i][j], total/nairs)
				}
				Dlog.Printf("   processFlow (%v, %v) airs: %v, total: %v\n", i, j, nairs, total)
			}
		}
	}
	tmp := *flow
	*flow = *flowBuffer
	*flowBuffer = tmp
	Dlog.Println("<- processFlow")
}
func (level *Level) Iterate(its int) {
	Dlog.Println("-> Level.Iterate")
	for it := 0; it < its; it++ {
		// Air
		level.processFlow(&level.air, &level.airBuffer,
			func(c Cell) bool { return c.AirFlows() },
			func(c Cell, a float64) float64 { return c.AirSinkSource(a) })
		// Energy
		level.processFlow(&level.energy, &level.energyBuffer,
			func(c Cell) bool { return c.EnergyFlows() },
			func(c Cell, a float64) float64 { return c.EnergySinkSource(a) })
	}
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

	sensor int
	energy_sensor_range int
	pressure_sensor_range int

	air_left, air_capacity float64
	helmet_on              bool

	copper, steel int
}

func (p *Player) Init() {
	p.x, p.y = 1, 1
	p.vision = 5

	p.energy_left, p.energy_capcacity = 1.0, 1.0

	p.sensor = noSensor
	p.pressure_sensor_range = 1
	p.energy_sensor_range = 1

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
	// Floor
	for i := 0; i < 31; i++ {
		for j := 0; j < 20; j++ {
			level.cells[i][j] = new(Floor)
		}
	}
	// Outer walls
	for i := 0; i < 31; i++ {
		level.cells[i][0] = new(Wall)
	}
	for i := 0; i < 31; i++ {
		level.cells[i][20] = new(Wall)
	}
	for i := 0; i < 20; i++ {
		level.cells[0][i] = new(Wall)
	}
	for i := 0; i < 20; i++ {
		level.cells[30][i] = new(Wall)
	}

	// Inner walls
	for i := 0; i < 15; i++ {
		level.cells[8][i] = new(Wall)
	}
	for i := 0; i < 18; i++ {
		level.cells[i][14] = new(Wall)
	}
	for i := 15; i < 20; i++ {
		level.cells[12][i] = new(Wall)
	}
	for i := 8; i < 18; i++ {
		level.cells[i][8] = new(Wall)
	}
	for i := 0; i < 20; i++ {
		level.cells[18][i] = new(Wall)
	}
	for i := 0; i < 20; i++ {
		level.cells[21][i] = new(Wall)
	}
	for i := 21; i < 30; i++ {
		level.cells[i][10] = new(Wall)
	}

	// Doors
	level.cells[0][2] = new(Door)
	level.cells[8][9] = new(Door)
	level.cells[14][8] = new(Door)
	level.cells[14][14] = new(Door)
	level.cells[12][17] = new(Door)
	level.cells[18][11] = new(Door)
	level.cells[21][6] = new(Door)
	level.cells[21][14] = new(Door)

	// Power Plant
	level.cells[26][3] = new(PowerPlant)
	level.cells[27][3] = new(PowerPlant)
	level.cells[26][4] = new(PowerPlant)
	level.cells[27][4] = new(PowerPlant)
/*
	level.cells[24][6].(*PowerPlant).damaged = false
	level.cells[25][6].(*PowerPlant).damaged = false
	level.cells[24][7].(*PowerPlant).damaged = false
	level.cells[25][7].(*PowerPlant).damaged = false
	*/

	// Air Plant
	level.cells[26][17] = new(AirPlant)
	level.cells[27][17] = new(AirPlant)
	level.cells[26][16] = new(AirPlant)
	level.cells[27][16] = new(AirPlant)

	// Conduits
	level.cells[28][4] = new(Conduit)
	level.cells[28][5] = new(Conduit)
	level.cells[28][6] = new(Conduit)
	level.cells[28][7] = new(Conduit)
	level.cells[28][8] = new(Conduit)
	level.cells[28][9] = new(Conduit)
	level.cells[28][10] = new(WallConduit)
	level.cells[28][11] = new(Conduit)
	level.cells[28][12] = new(Conduit)
	level.cells[28][13] = new(Conduit)
	level.cells[28][14] = new(Conduit)
	level.cells[28][15] = new(Conduit)
	level.cells[28][16] = new(Conduit)

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

	return game
}
func main() {
	file, err := os.Create("log")
	if err != nil {
		log.Fatal(err)
	}
	Dlog = log.New(file, "DERELICT: ", 0)

	game := NewGame()
	game.ui = NewCursesUI(&game.level, &game.player)
	game.ui.Run()
}
