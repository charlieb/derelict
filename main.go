package main

import "container/list"

type Level struct {
	items [][]*list.List
	cells [][]Cell
//	air   [][]Gas
}

type Cell interface {
	Walkable() bool
}
type Drawable interface {
	Character() string
}

type Player struct {
	pos [2]int
	items []Item
}

func (p *Player) Move(to [2]int, level *Level) {
	for e := level.items[p.pos[0]][p.pos[1]].Front(); e != nil; e = e.Next() {
		if e.Value == p {
			level.items[p.pos[0]][p.pos[1]].Remove(e)
			break
		}
	}
	p.pos = to
	level.items[p.pos[0]][p.pos[1]].PushFront(p)
}
func (p *Player) Walk(dir [2]int, level *Level) {
	pos := [2]int{p.pos[0] + dir[0], p.pos[1] + dir[1]}
	if level.cells[pos[0]][pos[1]].Walkable() {
		p.Move(pos, level)
	}
}
type Item interface {
	Activate(*Player)
}

type Wall struct {
}
func (w *Wall) Character() string { return "#" }

func main() {
	
}


