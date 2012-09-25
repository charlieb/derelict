package main

import (
	"container/list"
	"fmt"
	"github.com/errnoh/gocurse/curses"
	"math"
	"strings"
)

type UI interface {
	Run()
	Message(string)
	Menu(string, []string) (int, bool) // option, aborted
	DirectionPrompt() (int, int, bool) // x, y, abort
	YesNoPrompt(string) (bool, bool)   // Yes/No, aborted
}

const (
	none = iota
	revealMap
	airOverlay
	energyOverlay
	maxDebugMode
)

type CursesUI struct {
	screen   *curses.Window
	mapCache [][]int32
	messages list.List

	level  *Level
	player *Player

	debugMode int

	lookMode     bool
	lookX, lookY int
}

func NewCursesUI(level *Level, player *Player) UI {
	ui := new(CursesUI)
	ui.level = level
	ui.player = player
	ui.debugMode = none
	ui.lookMode = false

	// Init the mapCache to store seen parts of the level
	ui.mapCache = make([][]int32, level.x, level.x)
	for i := 0; i < level.x; i++ {
		ui.mapCache[i] = make([]int32, level.y, level.y)
		for j := 0; j < level.y; j++ {
			ui.mapCache[i][j] = ' '
		}
	}
	return ui
}
func (ui *CursesUI) Run() {
	// Initscr() initializes the terminal in curses mode.
	ui.screen, _ = curses.Initscr()
	// Endwin must be called when done.
	defer curses.Endwin()

	ui.setup()
	ui.drawMap()
	moved, quit := 0, false
	for !quit {
		// If any messages were added to the stack, draw them now
		if ui.messages.Len() > 0 {
			ui.drawMessages()
		}
		moved, quit = ui.handleKey(ui.screen.Getch())
		Dlog.Println("   RunCurses: ", moved, quit)
		if quit {
			ui.refresh()
			ui.messages.PushFront("Quit")
			ui.drawMessages()
			ui.screen.Getch()
			return
		}
		for it := 0; it < moved; it++ {
			ui.level.Iterate()
			ui.player.Iterate(ui.level)
		}
		ui.drawMap()
		ui.refresh()
	}
}
func (ui *CursesUI) Message(s string) { ui.messages.PushFront(s) }
func (ui *CursesUI) Menu(title string, s []string) (option int, aborted bool) {
	var (
		idx     int32 = 'a'
		max_len int   = len(title)
	)
	for i := 0; i < len(s); i++ {
		if len(s[i]) > max_len {
			max_len = len(s[i])
		}
	}
	max_len += 4 // 3 for "x: " 1 for the space after

	// All the i+1 below are due to the title line offset
	ui.screen.Addstr(0, 0, title, 0)
	for i := 0; i < len(s); i++ {
		if s[i] == "-" {
			ui.screen.Addstr(0, i+1, "  -------- ", 0)
			ui.screen.Addstr(11, i+1, strings.Repeat(" ", max_len-11), 0)
		} else {
			ui.screen.Addch(0, i+1, idx, 0)
			ui.screen.Addstr(1, i+1, ": ", 0)
			ui.screen.Addstr(3, i+1, s[i], 0)
			ui.screen.Addstr(3+len(s[i]), i+1, strings.Repeat(" ", max_len-(3+len(s[i]))), 0)
			idx++
		}
	}
	ui.screen.Addstr(0, len(s)+1, strings.Repeat(" ", max_len), 0)
	option = ui.screen.Getch() - 'a'
	if option < 0 || option >= int(idx) {
		aborted = true
	} else {
		aborted = false
	}
	// Interactive elements can call refresh to clear the screen
	ui.refresh()
	return
}
func (ui *CursesUI) drawMessages() {
	Dlog.Println("-> drawMessages")
	i := 0
	for e := ui.messages.Front(); e != nil; e = e.Next() {
		Dlog.Println("   drawMessages:", e.Value.(string))
		ui.screen.Addstr(0, i, e.Value.(string), 0)
		i++
	}
	ui.messages.Init()
	Dlog.Println("<- drawMessages")
}
func (ui *CursesUI) setup() {
	curses.Noecho()
	curses.Cbreak()
	ui.screen.Keypad(true)
	curses.Curs_set(0)
}
func (ui *CursesUI) YesNoPrompt(message string) (result, aborted bool) {
	ui.screen.Addstr(0, 0, message, 0)
	ui.screen.Addstr(len(message), 0, " Y/N ", 0)
	ch := ui.screen.Getch()
	if ch == 'y' || ch == 'Y' {
		return true, false
	} else if ch == 'n' || ch == 'N' {
		return false, false
	}
	return false, true
}
func castRay(x1, y1, x2, y2 int, cells [][]Cell) bool {
	Dlog.Println("-> castRay", x1, y1, x2, y2)
	var (
		dx, dy, ddx, ddy float64
	)
	dx = float64(x2 - x1)
	dy = float64(y2 - y1)

	// calc rise and tread 
	if dx == 0 && dy == 0 {
		return true
	}
	x1f, y1f := float64(x1), float64(y1)

	if math.Abs(dx) > math.Abs(dy) {
		if dx > 0 {
			ddx = 1
		} else {
			ddx = -1
		}
		if dy > 0 {
			ddy = math.Abs(dy / dx)
		} else {
			ddy = -math.Abs(dy / dx)
		}
		Dlog.Printf("   castRay: dx = %v, dy = %v, ddx = %v, ddy = %v", dx, dy, ddx, ddy)
		for i := 0.0; i < math.Abs(dx); i++ {
			cx, cy := int(x1f+i*ddx), int(y1f+i*ddy)
			Dlog.Printf("   castRay: %v, %v\n", cx, cy)
			if cells[cx][cy] != nil {
				if !cells[cx][cy].SeePast() {
					Dlog.Println("<- castRay", false)
					return false
				}
			}
		}
	} else {
		if dy > 0 {
			ddy = 1
		} else {
			ddy = -1
		}
		if dx > 0 {
			ddx = math.Abs(dx / dy)
		} else {
			ddx = -math.Abs(dx / dy)
		}
		Dlog.Printf("   castRay: dx = %v, dy = %v, ddx = %v, ddy = %v", dx, dy, ddx, ddy)
		for i := 0.0; i < math.Abs(dy); i++ {
			cx, cy := int(x1f+i*ddx), int(y1f+i*ddy)
			Dlog.Printf("   castRay: %v, %v\n", cx, cy)
			if cells[cx][cy] != nil {
				if !cells[cx][cy].SeePast() {
					Dlog.Println("<- castRay", false)
					return false
				}
			}
		}
	}
	Dlog.Println("<- castRay", true)
	return true
}
func drawSensor(rng, x, y, maxx, maxy int, sensed [][]float64, screen *curses.Window) {
	for i := -rng; i < rng; i++ {
		for j := -rng; j < rng; j++ {
			if i*i+j*j < rng*rng {
				if x+i >= 0 && x+i < maxx && y+j >= 0 && y+j < maxy {
					if sensed[x+i][y+j] >= 10 {
						screen.Addch(x+i, y+j, '9', 0)
					} else {
						screen.Addch(x+i, y+j, '0'+int32(sensed[x+i][y+j]), 0)
					}
				}
			}
		}
	}
}

func (ui *CursesUI) refresh() {
	var ch int32
	for i := 0; i < len(ui.mapCache); i++ {
		for j := 0; j < len(ui.mapCache[0]); j++ {
			switch ui.debugMode {
			case none:
				ch = ui.mapCache[i][j]
			case revealMap:
				ch = ui.level.cells[i][j].(Drawable).Character()
			case airOverlay:
				if ui.level.air[i][j] >= 10 {
					ch = '9'
				} else {
					ch = '0' + int32(ui.level.air[i][j])
				}
			case energyOverlay:
				if ui.level.energy[i][j] >= 10 {
					ch = '9'
				} else {
					ch = '0' + int32(ui.level.energy[i][j])
				}
			}
			ui.screen.Addch(i, j, ch, 0)
		}
	}

	// Draw player and sensor information if any
	switch ui.player.sensor {
	case noSensor:
		ui.screen.Addch(ui.player.x, ui.player.y, ui.player.Character(), 0)
	case pressureSensor:
		drawSensor(ui.player.pressure_sensor_range, ui.player.x, ui.player.y,
			ui.level.x, ui.level.y, ui.level.air, ui.screen)
	case energySensor:
		drawSensor(ui.player.energy_sensor_range, ui.player.x, ui.player.y,
			ui.level.x, ui.level.y, ui.level.air, ui.screen)
	}
	// Looking?
	if ui.lookMode {
		Dlog.Println("   refresh: lookMode")
		if ui.player.x == ui.lookX && ui.player.y == ui.lookY {
			ui.screen.Addch(ui.lookX, ui.lookY, ui.player.Character(), curses.A_REVERSE)
		} else {
			ui.screen.Addch(ui.lookX, ui.lookY, ui.mapCache[ui.lookX][ui.lookY], curses.A_REVERSE)
		}
	}
	ui.drawModeLine()
}

func (ui *CursesUI) drawMap() {
	Dlog.Println("-> CursesUI.drawMap")
	for i := -ui.player.vision; i < ui.player.vision; i++ {
		px := ui.player.x + i
		if px >= 0 && px < ui.level.x {
			for j := -ui.player.vision; j < ui.player.vision; j++ {
				py := ui.player.y + j
				if py >= 0 && py < ui.level.y {
					if i*i+j*j <= ui.player.vision*ui.player.vision {
						if castRay(ui.player.x, ui.player.y, px, py, ui.level.cells) {
							ui.mapCache[px][py] = ui.level.cells[px][py].(Drawable).Character()
							Dlog.Printf("   CurseUI.drawMap: %v %v drawn %c\n", px, py, ui.mapCache[px][py])
						}
					}
				}
			}
		}
	}
	ui.refresh()
	Dlog.Println("<- CursesUI.drawMap")
}
func (ui *CursesUI) drawModeLine() {
	var sensors string = "  "
	switch ui.player.sensor {
	case pressureSensor:
		sensors = "p"
	case energySensor:
		sensors = "e"
	}
	ui.screen.Addstr(0, 24, fmt.Sprintf("-- deReLict --  St:%v Cu:%v Air:%4.2f/%4.2f, Sensor:%v",
		ui.player.steel, ui.player.copper, ui.player.air_left,
		ui.player.air_capacity, sensors), 0)
}
func keyToDir(key int) (int, int, bool) { // dx,dy,abort
	switch key {
	case 'h':
		return -1, 0, false
	case 'l':
		return +1, 0, false
	case 'j':
		return 0, +1, false
	case 'k':
		return 0, -1, false

	case 'u':
		return +1, -1, false
	case 'y':
		return -1, -1, false
	case 'b':
		return -1, +1, false
	case 'n':
		return +1, +1, false
	case '.':
		return 0, 0, false
	}
	return 0, 0, true
}

func (ui *CursesUI) DirectionPrompt() (x, y int, abort bool) {
	ui.screen.Addstr(0, 0, "Which Direction?", 0)
	x, y, abort = keyToDir(ui.screen.Getch())
	ui.refresh()
	return
}
func (ui *CursesUI) handleKey(key int) (moved int, quit bool) {
	Dlog.Printf("-> handleKey key: %c", key)
	moved, quit = 0, false
	x, y, abort := keyToDir(key)
	if !abort {
		if ui.lookMode {
			if ui.lookX+x >= 0 && ui.lookX+x < ui.level.x &&
				ui.lookY+y >= 0 && ui.lookY+y < ui.level.y {
				ui.lookX += x
				ui.lookY += y
			}
			ui.Message(ui.level.cells[ui.lookX][ui.lookY].Description())
		} else if ui.player.Walk(x, y, ui.level) {
			moved = 1
		}
	} else {
		switch key {
		case 'm': // Action Menu
			moved = ui.player.Action(ui.level, ui, NONE)
		case 'c': // Create
			moved = ui.player.Action(ui.level, ui, CREATE)
		case 'r': // Repair
			moved = ui.player.Action(ui.level, ui, REPAIR)
		case 's': // Salvage
			moved = ui.player.Action(ui.level, ui, SALVAGE)
		case 'a': // Activate
			moved = ui.player.Action(ui.level, ui, ACTIVATE)
		case 'p': // Toggle Pressure Sensor
			if ui.player.sensor == pressureSensor {
				ui.player.sensor = noSensor
			} else {
				ui.player.sensor = pressureSensor
			}
			ui.refresh()
		case 'e': // Toggle Energy Sensor
			if ui.player.sensor == energySensor {
				ui.player.sensor = noSensor
			} else {
				ui.player.sensor = energySensor
			}
			ui.refresh()
		case ';': // Toggle look mode
			ui.lookMode = !ui.lookMode
			if ui.lookMode {
				ui.lookX = ui.player.x
				ui.lookY = ui.player.y
				ui.Message("Looking around - this is you")
			}
		case 'd': // Debug
			ui.debugMode++
			if ui.debugMode == maxDebugMode {
				ui.debugMode = none
			}
		case 'q':
			quit = true
		}
	}
	Dlog.Printf("<- handleKey moved: %v, quit: %v", moved, quit)
	return
}
