package main

type Level struct {
	cells [][][]Item
}

type Player struct {
	pos [2]int
	items []Item
}

func (p *Player) Move(to [2]int) {
	p.pos = to
}
type Item interface {
	Activate(*Player)
	Walkable() bool
	Character() string
}

func main() {
	
}


