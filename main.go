package main

import (
	"bytes"
	_ "embed"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bendahl/uinput"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/gltext"
)

type Mode string

type Cell struct {
	x      int
	y      int
	label  string
	screen *Screen
}

type Screen struct {
	win   *glfw.Window
	font  *gltext.Font
	cells []*Cell
}

type State struct {
	input       []byte
	mode        Mode
	mouse       uinput.Mouse
	screens     []*Screen
	cells       map[string]*Cell
	focusedCell *Cell
}

const cellSize = 120

const GridMode Mode = "GridMode"

const CursorMode Mode = "CursorMode"

//go:embed JetBrainsMono-Bold.ttf
var embededFont []byte

var letters = [...]string{
	"a", "s", "d", "f", "g", "h", "j", "k", "l",
	"q", "w", "e", "r", "t", "y", "u", "i", "o", "p",
	"z", "x", "c", "v", "b", "n", "m"}

var generatedLabels = make(map[string]bool)

var state *State

func createScreen(m *glfw.Monitor) *Screen {
	mode := m.GetVideoMode()

	win, err := glfw.CreateWindow(mode.Width, mode.Height, m.GetName(), nil, nil)
	if err != nil {
		panic(err)
	}

	mx, my := m.GetPos()
	win.SetPos(mx, my)

	screen := &Screen{
		win,
		nil,
		nil,
	}

	screen.cells = makeCells(mode.Width/cellSize, mode.Height/cellSize, screen)
	return screen
}

func makeCells(cols int, rows int, screen *Screen) []*Cell {
	cells := make([]*Cell, 0)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			label := makeLabel()
			cell := &Cell{c * cellSize, r * cellSize, label, screen}
			state.cells[label] = cell
			cells = append(cells, cell)
		}
	}

	return cells
}

func makeLabel() string {
	c := len(letters)
	rand.Seed(time.Now().UnixNano())

	for {
		l1 := rand.Intn(c)
		l2 := rand.Intn(c)
		label := letters[l1] + letters[l2]

		if !generatedLabels[label] {
			generatedLabels[label] = true
			return strings.ToUpper(label)
		}
	}
}

func (s *Screen) render() {
	s.win.MakeContextCurrent()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(0, 0, 0, 0)

	if state.mode == CursorMode {
		s.win.SwapBuffers()
		return
	}

	for _, c := range s.cells {
		cs := float32(cellSize)
		hcs := float32(cellSize / 2)
		x := float32(c.x)
		y := float32(c.y)
		isMatch := len(state.input) > 0 && state.input[0] == c.label[0]

		if !isMatch {
			gl.Color4f(0, 0, 0, 0.2)
			gl.Rectf(x+1, y+1, x+cs-1, y+cs-1)
		}

		lw, lh := s.font.Metrics(c.label)

		if isMatch {
			gl.Color4f(0, 0, 0, 0.2)
		} else {
			gl.Color4f(0, 0, 0, 0.4)
		}

		gl.Rectf(x+hcs-float32(lw), y+hcs-float32(lh/2), x+hcs+float32(lw), y+hcs+float32(lh/2))

		gl.Color4f(0, 0, 0, 1)
		s.font.Printf(x+hcs-float32(lw/2), y+hcs-float32(lh/2), c.label)

		gl.Color4f(1, 0.8, 0, 1)
		s.font.Printf(x+hcs-float32(lw/2), y+hcs-float32(lh/2), c.label)

		if isMatch {
			gl.Color4f(1, 0.0, 0, 1)
			s.font.Printf(x+hcs-float32(lw/2), y+hcs-float32(lh/2), string(c.label[0]))
		}
	}

	s.win.SwapBuffers()
}

func loadFont() *gltext.Font {
	font, err := gltext.LoadTruetype(bytes.NewReader(embededFont), 16, 32, 127, gltext.LeftToRight)
	if err != nil {
		panic(err)
	}

	return font
}

func initGlfw() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.FocusOnShow, glfw.True)
	glfw.WindowHint(glfw.Focused, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Floating, glfw.True)
	glfw.WindowHint(glfw.Decorated, glfw.True)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
}

func initGl() {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.LIGHTING)
	gl.Enable(gl.DEPTH_TEST)

	gl.Enable(gl.TEXTURE_2D)
	gl.DepthFunc(gl.LEQUAL)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	for _, s := range state.screens {
		s.win.MakeContextCurrent()
		glfw.SwapInterval(0)

		sw, sh := s.win.GetSize()

		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()
		gl.Ortho(0, float64(sw), float64(sh), 0, 0, 1)

		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()
		gl.ClearColor(0, 0, 0, 1)

		s.font = loadFont()
	}
}

func onKey(win *glfw.Window, key glfw.Key, _ int, action glfw.Action, mods glfw.ModifierKey) {
	win.Focus()
	win.SwapBuffers()

	if key == glfw.KeyEscape || key == glfw.KeyCapsLock {
		cleanUp()
		return
	}

	if state.mode == GridMode {
		if key >= 65 && key <= 90 && action == glfw.Release {
			state.input = append(state.input, byte(key))

			if len(state.input) >= 3 {
				state.input = state.input[len(state.input)-2:]
			}
		}

		if key == glfw.KeyBackspace && len(state.input) > 0 {
			state.input = state.input[:len(state.input)-1]
		}

		if cell, ok := state.cells[string(state.input)]; ok {
			state.mode = CursorMode
			padding := cellSize / 2
			state.focusedCell = cell
			cell.screen.win.Focus()
			cell.screen.win.SwapBuffers()
			cell.screen.win.SetCursorPos(float64(cell.x+padding), float64(cell.y+padding))
		}
	}

	if state.mode == CursorMode {
		m := state.mouse
		step := int32(3)

		if mods&glfw.ModControl != 0 {
			step = 15
		}

		if key == glfw.KeyH {
			m.MoveLeft(step)
		}

		if key == glfw.KeyJ {
			m.MoveDown(step)
		}

		if key == glfw.KeyK {
			m.MoveUp(step)
		}

		if key == glfw.KeyL {
			m.MoveRight(step)
		}

		if key == glfw.KeyEnter {
			state.focusedCell.screen.win.Hide()
			state.focusedCell.screen.win.SwapBuffers()

			if mods&glfw.ModShift != 0 {
				m.RightClick()
			} else if mods&glfw.ModControl != 0 {
				m.LeftClick()
				m.LeftClick()
			} else {
				m.LeftClick()
			}

			cleanUp()
		}
	}
}

func initScreens() {
	for _, m := range glfw.GetMonitors() {
		s := createScreen(m)
		s.win.SetKeyCallback(onKey)

		state.screens = append(state.screens, s)
	}
}

func createMouse() uinput.Mouse {
	dev, err := uinput.CreateMouse("/dev/uinput", []byte("ratazana"))
	if err != nil {
		panic(err)
	}

	return dev
}

func runLoop() {
	for {
		for _, s := range state.screens {
			if s.win.ShouldClose() {
				return
			}

			s.render()
		}

		glfw.PollEvents()
	}
}

func cleanUp() {
	for _, s := range state.screens {
		s.font.Release()
		state.mouse.Close()
		s.win.Destroy()
	}

	glfw.Terminate()
	os.Exit(0)
}

func main() {
	runtime.LockOSThread()

	state = &State{
		input:   make([]byte, 0),
		mode:    GridMode,
		mouse:   createMouse(),
		screens: make([]*Screen, 0),
		cells:   make(map[string]*Cell),
	}

	initGlfw()
	initScreens()
	initGl()
	runLoop()
	cleanUp()
}
