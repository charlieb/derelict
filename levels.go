package main

import (
	"fmt"
	"math/rand"
  "os"
  "log"
)

func GenerateLevel() {}

const (
	NORMAL = iota
	CORRIDOR
)
type Room struct { x, y, w, h, kind int }

func (r *Room) Print() {
	fmt.Printf("%v, %v - %v, %v\n", r.x, r.y, r.w, r.h)
}

func subdiv(room *Room) (bool, []*Room) {
	rooms := make([]*Room, 2)
	// Choose an axis
	if 0 == rand.Intn(2) {
		if room.w <= 5 { // #.#.# 
			return false, rooms
		}
		neww := 3 + rand.Intn(room.w - 5)
		rooms[0] = &Room{ room.x, room.y, neww, room.h, room.kind }
		rooms[1] = &Room{ room.x + neww - 1, room.y, room.w - neww + 1, room.h, room.kind }
	} else {
		if room.h <= 5 {
			return false, rooms
		}
		newh := 3 + rand.Intn(room.h - 5)
		rooms[0] = &Room{ room.x, room.y, room.w, newh, room.kind }
		rooms[1] = &Room{ room.x, room.y + newh - 1, room.w, room.h - newh + 1, room.kind }
	}
	return true, rooms
}

func roomsToLevel(rooms []*Room) *Level {
  level := new(Level)
	level.x, level.y = 69, 23
	level.Init()

  for _, r := range rooms {
    for i := r.x; i < r.x + r.w; i++ {
      level.cells[i][r.y] = new(Wall)
      level.cells[i][r.y + r.h - 1] = new(Wall)
    }
    for j := r.y; j < r.y + r.h; j++ {
      level.cells[r.x][j] = new(Wall)
      level.cells[r.x + r.w - 1][j] = new(Wall)
    }
  }
  return level
}

type RoomNode struct { room *Room, rooms []*Room }



func testLevel() {
	r := Room{0,0,69,23, NORMAL}
	r.Print()
	fmt.Println("---")
	_, rooms := subdiv(&r)
	for i := 0; i < 10; i++ {
		new_rooms := make([]*Room, 0)
		for _, r := range rooms {
			ok, rs := subdiv(r)
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

	file, err := os.Create("log")
	if err != nil {
		log.Fatal(err)
	}
	Dlog = log.New(file, "DERELICT: ", 0)

  player := new(Player)
  player.Init()
  ui := NewCursesUI(roomsToLevel(rooms), player)
	ui.Run()

}

