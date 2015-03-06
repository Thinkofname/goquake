package main

import (
	"fmt"
	"github.com/davecheney/profile"
	"github.com/go-gl/glfw/v3.0/glfw"
	"github.com/thinkofdeath/goquake/bsp"
	"github.com/thinkofdeath/goquake/pak"
	"github.com/thinkofdeath/goquake/render"
	"runtime"
	"time"
)

func init() {
	runtime.LockOSThread()
}

var (
	lockMouse = true
)

func main() {
	defer profile.Start(&profile.Config{
		CPUProfile:  true,
		ProfilePath: "./profiles",
	}).Stop()

	if !glfw.Init() {
		panic("glfw error")
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

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
	defer p.Close()
	bsp, err := bsp.ParseBSPFile(p.Reader("maps/start.bsp"))
	if err != nil {
		panic(err)
	}

	render.Init(p, bsp)

	fmt.Println(time.Now().Sub(start))

	window.SetKeyCallback(onKey)
	window.SetCursorPositionCallback(onMouseMove)
	window.SetInputMode(glfw.Cursor, glfw.CursorHidden)
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
		(xpos-ww)/3000,
		(ypos-hh)/3000,
	)
}

func onMouse(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft && action == glfw.Press {
		lockMouse = true
		w.SetInputMode(glfw.Cursor, glfw.CursorHidden)
	}
}
