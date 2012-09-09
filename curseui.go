package main

import (
	"container/list"
	"github.com/errnoh/gocurse/curses"
	"math"
	"strings"
)

type UI interface {
	Run(*Game)
	Message(string)
	Menu(string, []string) (int, bool) // option, aborted
	DirectionPrompt() (int, int, bool) // x, y, abort
	YesNoPrompt(string) (bool, bool)   // Yes/No, aborted
}
type CursesUI struct {
	screen   *curses.Window
	mapCache [][]int32
	messages list.List
}

func (ui *CursesUI) Run(game *Game) {
	// Initscr() initializes the terminal in curses mode.
	ui.screen, _ = curses.Initscr()
	// Endwin must be called when done.
	defer curses.Endwin()

	ui.setup()

	// Init the mapCache to store seen parts of the level
	ui.mapCache = make([][]int32, game.level.x, game.level.x)
	for i := 0; i < game.level.x; i++ {
		ui.mapCache[i] = make([]int32, game.level.y, game.level.y)
		for j := 0; j < game.level.y; j++ {
			ui.mapCache[i][j] = ' '
		}
	}

	ui.drawMap(&game.level, &game.player)
	for {
		moved, quit := 0, false
		for moved <= 0 {
			// If any messages were added to the stack, draw them now
			if ui.messages.Len() > 0 {
				ui.drawMessages()
			}
			moved, quit = ui.handleKey(ui.screen.Getch(), &game.level, &game.player)
			Dlog.Println("   RunCurses: ", moved, quit)
			if quit {
				ui.messages.PushFront("Quit")
				ui.drawMessages()
				ui.screen.Getch()
				return
			}
		}
		game.level.Iterate(moved)
		ui.drawMap(&game.level, &game.player)
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
		ddx = 1
		ddy = dy / dx
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
		ddx = dx / dy
		if dy > 0 {
			ddy = 1
		} else {
			ddy = -1
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
func (ui *CursesUI) refresh() {
	for i := 0; i < len(ui.mapCache); i++ {
		for j := 0; j < len(ui.mapCache[0]); j++ {
			ui.screen.Addch(i, j, ui.mapCache[i][j], 0)
		}
	}
}

func (ui *CursesUI) drawMap(level *Level, player *Player) {
	Dlog.Println("-> CursesUI.drawMap")
	for i := -player.vision; i < player.vision; i++ {
		px := player.x + i
		if px >= 0 && px < level.x {
			for j := -player.vision; j < player.vision; j++ {
				py := player.y + j
				if py >= 0 && py < level.y {
					if i*i+j*j <= player.vision*player.vision {
						if castRay(player.x, player.y, px, py, level.cells) {
							ui.mapCache[px][py] = '.'
							if level.cells[px][py] != nil {
								ui.mapCache[px][py] = level.cells[px][py].(Drawable).Character()
							}
						}
					}
				}
			}
		}
	}
	ui.mapCache[player.x][player.y] = player.Character()
	ui.refresh()
	Dlog.Println("<- CursesUI.drawMap")
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
func (ui *CursesUI) handleKey(key int, level *Level, player *Player) (moved int, quit bool) {
	Dlog.Printf("-> handleKey key: %c", key)
	moved, quit = 0, false
	x, y, abort := keyToDir(key)
	if !abort {
		if player.Walk(x, y, level) {
			moved = 1
		}
	} else {
		switch key {
		case 'm': // Action Menu
			moved = player.Action(level, ui, NONE)
		case 'c': // Create
			moved = player.Action(level, ui, CREATE)
		case 'r': // Repair
			moved = player.Action(level, ui, REPAIR)
		case 's': // Salvage
			moved = player.Action(level, ui, SALVAGE)
		case 'a': // Activate
			moved = player.Action(level, ui, ACTIVATE)
		case 'q':
			quit = true
		}
	}
	Dlog.Printf("<- handleKey moved: %v, quit: %v", moved, quit)
	return
}
