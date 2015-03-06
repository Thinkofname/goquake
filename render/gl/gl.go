// Package gl provides a more Go friendly OpenGL API
package gl

import (
	"fmt"
	"github.com/go-gl/gl/v2.1/gl"
)

const (
	ColorBufferBit ClearFlags = gl.COLOR_BUFFER_BIT
	DepthBufferBit ClearFlags = gl.DEPTH_BUFFER_BIT

	DepthTest    Flag = gl.DEPTH_TEST
	CullFaceFlag Flag = gl.CULL_FACE

	Back  Face = gl.BACK
	Front Face = gl.FRONT

	ClockWise        FaceDirection = gl.CW
	CounterClockWise FaceDirection = gl.CCW

	Triangles DrawType = gl.TRIANGLES
)

func Init() {
	if err := gl.Init(); err != nil {
		panic(err)
	}
}

type (
	ClearFlags    uint32
	Flag          uint32
	Face          uint32
	FaceDirection uint32
	DrawType      uint32
)

func Viewport(x, y, width, height int) {
	gl.Viewport(int32(x), int32(y), int32(width), int32(height))
}

func ClearColor(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
}

func Clear(flags ClearFlags) {
	gl.Clear(uint32(flags))
}

func ActiveTexture(id int) {
	gl.ActiveTexture(gl.TEXTURE0 + uint32(id))
}

func Enable(flag Flag) {
	gl.Enable(uint32(flag))
}

func Disable(flag Flag) {
	gl.Disable(uint32(flag))
}

func CullFace(face Face) {
	gl.CullFace(uint32(face))
}

func FrontFace(dir FaceDirection) {
	gl.FrontFace(uint32(dir))
}

func DrawArrays(ty DrawType, offset, count int) {
	gl.DrawArrays(uint32(ty), int32(offset), int32(count))
}

func checkError() {
	err := gl.GetError()
	if err != 0 {
		panic(fmt.Sprintf("gl error: %d", err))
	}
}

func Flush() {
	gl.Flush()
}
