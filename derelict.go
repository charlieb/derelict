package main

import (
	"container/list"
)

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

type Cell interface {
	Walkable() bool
}
type Drawable interface {
	Character() int32
}

////////////////////// PLAYER /////////////////////////

type Player struct {
	pos [2]int
	items []Item
}

func (p *Player) Move(to [2]int, level *Level) {
//	for e := level.items[p.pos[0]][p.pos[1]].Front(); e != nil; e = e.Next() {
//		if e.Value == p {
//			level.items[p.pos[0]][p.pos[1]].Remove(e)
//			break
//		}
//	}
	p.pos = to
//	level.items[p.pos[0]][p.pos[1]].PushFront(p)
}
func (p *Player) Walk(dir [2]int, level *Level) {
	pos := [2]int{p.pos[0] + dir[0], p.pos[1] + dir[1]}
	if level.cells[pos[0]][pos[1]].Walkable() {
		p.Move(pos, level)
	}
}
func (p *Player) Character() int32 { return '@' }

////////////////////// /////////////////////////

type Item interface {
	Activate(*Player)
}

type Wall struct {
}
func (w *Wall) Character() int32 { return '#' }

func main() {
	var level Level
	level.x, level.y = 69, 23
	level.Init()
	RunCursesUI(&level)
}


