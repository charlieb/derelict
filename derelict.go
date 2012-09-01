package main

import (
	"container/list"
	"log"
	"os"
)

var dlog *log.Logger
////////////////////// LEVEL /////////////////////////

type Level struct {
	x, y int
	items [][]*list.List
	cells [][]Cell
//	air   [][]Gas
}
func (level *Level) Init() {
	level.items = make([][]*list.List, level.x, level.x)
	level.cells = make([][]Cell, level.x, level.x)
	for i := 0; i < level.x; i++ {
		level.cells[i] = make([]Cell, level.y, level.y)
		level.items[i] = make([]*list.List, level.y, level.y)
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
		level.cells[i][level.y - 1] = new(Wall)
	}
	for j := 0; j < level.y; j++ {
		level.cells[0][j] = new(Wall)
	}
	for j := 0; j < level.y; j++ {
		level.cells[level.x - 1][j] = new(Wall)
	}
}

type Cell interface {
	Walkable() bool

	SeePast() bool
}
type Item interface {
	Activate(*Player)
}
type Drawable interface {
	Character() int32
}

////////////////////// PLAYER /////////////////////////

type Player struct {
	x, y int
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
	dlog.Println("Move", p.x, p.y)
	p.x, p.y = to_x, to_y
//	level.items[p.pos[0]][p.pos[1]].PushFront(p)
}
func (p *Player) Walk(dir_x, dir_y int, level *Level) {
	px, py := p.x + dir_x, p.y + dir_y
	dlog.Println("Walk", px, py)
	if px >= 0 && px < level.x && py >= 0 && py < level.y {
		if level.cells[px][py] == nil || level.cells[px][py].Walkable() {
			p.Move(px, py)
		}
	}
}
func (p *Player) Character() int32 { return '@' }

////////////////////// /////////////////////////


type sCell struct {
	visible bool
}
func (c *sCell) Walkable() bool { return true }
func (c *sCell) SeePast() bool { return true }

type Wall struct {
	sCell
}
func (w *Wall) Walkable() bool { return false }
func (w *Wall) Character() int32 { return '#' }
func (w *Wall) SeePast() bool { return false }

func buildTestWalls(level *Level) {
	for i := 2; i < 20; i++ {
		level.cells[i][2] = new(Wall)
	}
	for j := 3; j < 20; j++ {
		level.cells[2][j] = new(Wall)
	}
}

func main() {
	file, err := os.Create("log")
	if err != nil { log.Fatal(err) }
	dlog = log.New(file, "DERELICT: ", 0)

	var level Level
	level.x, level.y = 69, 23
	level.Init()
	level.outerWall()

	buildTestWalls(&level)

	var player Player
	player.x, player.y, player.vision = 1,1,5
	RunCursesUI(&level, &player)
}


