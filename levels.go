package main

import (
	"fmt"
	"math/rand"
  "os"
  "log"
)

type Room interface {
  Print()
  minX() int
  minY() int
  maxX() int
  maxY() int
  addToLevel(*Level)
}

func GenerateLevel() {}


///////////// RECTANGLE ROOM ///////////////////
type RectRoom struct { x, y, w, h int }
func (r *RectRoom) Print() {
  fmt.Printf("%v, %v -> %v, %v\n", r.minX(), r.minY(), r.maxX(), r.maxY())
}
func (r *RectRoom) minX() int { return r.x }
func (r *RectRoom) minY() int { return r.y }
func (r *RectRoom) maxX() int { return r.x + r.w - 1 }
func (r *RectRoom) maxY() int { return r.y + r.h - 1 }
func (room *RectRoom) addToLevel(level *Level) {
  for i := room.x + 1; i < room.x + room.w -1; i++ {
    for j := room.y + 1; j < room.y + room.h -1; j++ {
      level.cells[i][j] = new(Floor)
    }
  }
  for i := room.x; i < room.x + room.w; i++ {
    level.cells[i][room.y] = new(Wall)
    level.cells[i][room.y + room.h - 1] = new(Wall)
  }
  for j := room.y; j < room.y + room.h; j++ {
    level.cells[room.x][j] = new(Wall)
    level.cells[room.x + room.w - 1][j] = new(Wall)
  }
}
func (room *RectRoom) subdiv() (bool, []*RectRoom) {
	rooms := make([]*RectRoom, 2)
	// Choose an axis
	if 0 == rand.Intn(2) {
		if room.w <= 5 { // #.#.# 
			return false, rooms
		}
		neww := 3 + rand.Intn(room.w - 5)
		rooms[0] = &RectRoom{ room.x, room.y, neww, room.h }
		rooms[1] = &RectRoom{ room.x + neww - 1, room.y, room.w - neww + 1, room.h }
	} else {
		if room.h <= 5 {
			return false, rooms
		}
		newh := 3 + rand.Intn(room.h - 5)
		rooms[0] = &RectRoom{ room.x, room.y, room.w, newh }
		rooms[1] = &RectRoom{ room.x, room.y + newh - 1, room.w, room.h - newh + 1 }
	}
	return true, rooms
}

///////////// CIRCLE ROOM ///////////////////
type CircleRoom struct { x,y,r int }
func (r *CircleRoom) Print() {
  fmt.Printf("%v, %v -> %v, %v\n", r.minX(), r.minY(), r.maxX(), r.maxY())
}
func (r *CircleRoom) minX() int { return r.x - r.r }
func (r *CircleRoom) minY() int { return r.y - r.r }
func (r *CircleRoom) maxX() int { return r.x + r.r }
func (r *CircleRoom) maxY() int { return r.y + r.r }
func (r *CircleRoom) addToLevel(level *Level) {
  x := r.r
  y := 0;
  prev_x := x
  prev_y := y
  radiusError := 1 - x;
  octate := func(cx,cy,rx,ry int) { //cx = center, rx = radius
    fmt.Printf("%v, %v\n", cx + rx, cy + ry)
    level.cells[cx + rx][cy + ry] = new(Wall)
    level.cells[cx + ry][cy + rx] = new(Wall)
    level.cells[cx - rx][cy + ry] = new(Wall)
    level.cells[cx - ry][cy + rx] = new(Wall)
    level.cells[cx - rx][cy - ry] = new(Wall)
    level.cells[cx - ry][cy - rx] = new(Wall)
    level.cells[cx + rx][cy - ry] = new(Wall)
    level.cells[cx + ry][cy - rx] = new(Wall)
    for i := 0; i < rx; i++ {
      for j := 0; j < ry; j++ {
        level.cells[cx + i][cy + j] = new(Floor)
        level.cells[cx + j][cy + i] = new(Floor)
        level.cells[cx - i][cy + j] = new(Floor)
        level.cells[cx - j][cy + i] = new(Floor)
        level.cells[cx - i][cy - j] = new(Floor)
        level.cells[cx - j][cy - i] = new(Floor)
        level.cells[cx + i][cy - j] = new(Floor)
        level.cells[cx + j][cy - i] = new(Floor)
      }
    }
  }

  for x >= y {
    octate(r.x,r.y, x,y)
    if x != prev_x && y != prev_y {
      fmt.Println('-')
      octate(r.x,r.y,x,prev_y)
    }
    prev_x = x
    prev_y = y

    y++;
    if radiusError < 0  {
      radiusError += 2 * y + 1;
    } else {
      x--;
      radiusError += 2 * (y - x) + 1;
    }
  }
  octate(r.x,r.y,x,prev_y)
}


///////////// LEVEL GEN ///////////////////
func min(x, y int) int { if x < y { return x } else { return y };return 0}
func generateLevel(x, y int) *Level {
  level := new(Level)
	level.x, level.y = x,y
	level.Init()
  for i := 0; i < 50; i++ {
    room := new(RectRoom)
    room.x = rand.Intn(x-5)
    room.y = rand.Intn(y-5)
    room.w = 5 + rand.Intn(min(10,x - room.x - 5))
    room.h = 5 + rand.Intn(min(10,y - room.y - 5))
    room.addToLevel(level)
  }
  for i := 0; i < 5; i++ {
    room := new(CircleRoom)
    room.x = 5 + rand.Intn(x-10)
    room.y = 5 + rand.Intn(y-10)
    room.r = 2 + rand.Intn(5)
    room.addToLevel(level)
  }
  return level
}



func testLevel() {
	r := RectRoom{0,0,69,23}
	r.Print()
	fmt.Println("---")
	_, rooms := r.subdiv()
	for i := 0; i < 10; i++ {
		new_rooms := make([]*RectRoom, 0)
		for _, r := range rooms {
			ok, rs := r.subdiv()
			if ok {
				new_rooms = append(new_rooms, rs...)
			} else {
				new_rooms = append(new_rooms, r)
			}
		}
    for _,r := range new_rooms { r.Print()}
		fmt.Println("---")
		rooms = new_rooms
	}

	file, err := os.Create("level")
	if err != nil {
		log.Fatal(err)
	}
	Dlog = log.New(file, "DERELICT: ", 0)

  /*
  level := new(Level)
	level.x, level.y = 69, 23
	level.Init()
  for _, r := range rooms { r.addToLevel(level) }
  cr := new(CircleRoom)
  cr.x, cr.y, cr.r = 20 + rand.Intn(20),10,rand.Intn(9)
  cr.addToLevel(level)
  */

  level := generateLevel(69, 23)

  player := new(Player)
  player.Init()
  ui := NewCursesUI(level, player)
	ui.Run()

}

