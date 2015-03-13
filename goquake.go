package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.0/glfw"
	"github.com/thinkofdeath/goquake/pak"
	"github.com/thinkofdeath/goquake/render"
	"math/rand"
	"runtime"
	"time"
)

func init() {
	runtime.LockOSThread()
}

var (
	lockMouse = false
)

func main() {
	if !glfw.Init() {
		panic("glfw error")
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
	glfw.WindowHint(glfw.OpenglForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(640, 480, "GoQuake", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	glfw.SwapInterval(1)

	start := time.Now()
	p, err := pak.FromFile("id1/PAK0.PAK")
	if err != nil {
		panic(err)
	}

	// Check for the full game
	p2, err := pak.FromFile("id1/PAK1.PAK")
	if err == nil {
		p = pak.Join(p, p2)
	}

	defer p.Close()

	render.Init(p)

	fmt.Println(time.Now().Sub(start))

	window.SetKeyCallback(onKey)
	window.SetCursorPositionCallback(onMouseMove)
	window.SetMouseButtonCallback(onMouse)

	for !window.ShouldClose() {
		width, height := window.GetFramebufferSize()

		render.Draw(width, height)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func onKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyW {
		switch action {
		case glfw.Release:
			render.StopMove()
		case glfw.Press:
			render.MoveForward()
		}
	} else if key == glfw.KeyEscape {
		lockMouse = false
		w.SetInputMode(glfw.Cursor, glfw.CursorNormal)
	} else if key == glfw.Key1 && action == glfw.Release {
		levels := []string{
			"b_bh10",
			"b_bh100",
			"b_bh25",
			"b_shell1",
			"b_shell0",
			"b_nail1",
			"b_nail0",
			"b_rock1",
			"b_rock0",
			"b_batt1",
			"b_batt0",
			"b_explob",
			"start",
			"e1m1",
			"e1m2",
			"e1m3",
			"e1m4",
			"e1m5",
			"e1m6",
			"e1m7",
			"e1m8",
			"b_exbox2",
			"e2m1",
			"e2m2",
			"e2m3",
			"e2m4",
			"e2m5",
			"e2m6",
			"e2m7",
			"e3m1",
			"e3m2",
			"e3m3",
			"e3m4",
			"e3m5",
			"e3m6",
			"e3m7",
			"e4m1",
			"e4m2",
			"e4m3",
			"e4m4",
			"e4m5",
			"e4m6",
			"e4m7",
			"e4m8",
			"end",
			"dm1",
			"dm2",
			"dm3",
			"dm4",
			"dm5",
			"dm6",
		}

		render.SetLevel(levels[rand.Intn(len(levels))])
	}
}

func onMouseMove(w *glfw.Window, xpos float64, ypos float64) {
	if !lockMouse {
		return
	}
	width, height := w.GetFramebufferSize()
	ww, hh := float64(width/2), float64(height/2)
	w.SetCursorPosition(ww, hh)

	render.Rotate(
		(xpos-ww)/2000,
		(ypos-hh)/2000,
	)
}

func onMouse(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft && action == glfw.Press {
		lockMouse = true
		w.SetInputMode(glfw.Cursor, glfw.CursorDisabled)
	}
}
