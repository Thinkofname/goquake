package main

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.0/glfw"
	"github.com/thinkofdeath/goquake/bsp"
	"github.com/thinkofdeath/goquake/pak"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func main() {
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

	if err := gl.Init(); err != nil {
		panic(err)
	}

	glfw.SwapInterval(1)

	p, err := pak.FromFile("id1/PAK0.PAK")
	if err != nil {
		panic(err)
	}
	defer p.Close()
	bsp, err := bsp.ParseBSPFile(p.Reader("maps/start.bsp"))
	if err != nil {
		panic(err)
	}
	_ = bsp

	for !window.ShouldClose() {
		width, height := window.GetFramebufferSize()
		gl.Viewport(0, 0, int32(width), int32(height))
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
