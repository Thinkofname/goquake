package gl

import (
	"github.com/go-gl/gl/v2.1/gl"
)

const (
	ArrayBuffer BufferTarget = gl.ARRAY_BUFFER

	StaticDraw  BufferUsage = gl.STATIC_DRAW
	DynamicDraw BufferUsage = gl.DYNAMIC_DRAW
)

type Buffer struct {
	internal uint32
}

type BufferTarget uint32
type BufferUsage uint32

func CreateBuffer() Buffer {
	var buffer Buffer
	gl.GenBuffers(1, &buffer.internal)
	return buffer
}

var (
	currentBuffer       Buffer
	currentBufferTarget BufferTarget
)

func (b Buffer) Bind(target BufferTarget) {
	if currentBuffer == b && currentBufferTarget == target {
		return
	}
	gl.BindBuffer(uint32(target), b.internal)
	currentBuffer = b
	currentBufferTarget = target
}

func (b Buffer) DataFloat32(data []float32, usage BufferUsage) {
	if currentBuffer != b {
		panic("buffer not bound")
	}
	gl.BufferData(uint32(currentBufferTarget), len(data)*4, gl.Ptr(data), uint32(usage))
}
