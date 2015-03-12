package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.0/glfw"
	"github.com/thinkofdeath/goquake/pak"
	"github.com/thinkofdeath/goquake/render"
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
		render.SetLevel("e1m1")
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
