// Package builder provides a simple way to create buffers to upload to the gpu
package builder

import (
	"bytes"
	"encoding/binary"
	"math"
	"unsafe"
)

// Types allowed to be used in a buffer.
const (
	UnsignedByte  Type = 1
	Byte          Type = 1
	UnsignedShort Type = 2
	Short         Type = 2
	Float         Type = 4
)

var nativeOrder = func() binary.ByteOrder {
	check := uint32(1)
	c := (*[4]byte)(unsafe.Pointer(&check))
	if binary.LittleEndian.Uint32(c[:]) == 1 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}()

// Type is a type that is allowed in a buffer.
type Type int

// Buffer is a dynamically sized byte buffer
// for creating data to upload to the gpu.
type Buffer struct {
	buf      bytes.Buffer
	elemSize int

	scratch [8]byte
}

// Creates a new buffer containing the passed
// types.
func New(types ...Type) *Buffer {
	elemSize := 0
	for _, t := range types {
		elemSize += int(t)
	}
	b := &Buffer{
		elemSize: elemSize,
	}
	b.buf.Grow(elemSize * 300)
	return b
}

// UnsignedByte appends an unsigned byte to the
// buffer.
func (b *Buffer) UnsignedByte(i byte) {
	b.buf.WriteByte(i)
}

// Byte appends a signed byte to the buffer.
func (b *Buffer) Byte(i int8) {
	b.UnsignedByte(byte(i))
}

// UnsignedShort writes an unsigned short to the
// buffer
func (b *Buffer) UnsignedShort(i uint16) {
	d := b.scratch[:2]
	nativeOrder.PutUint16(d, i)
	b.buf.Write(d)
}

// Short writes a short to the buffer.
func (b *Buffer) Short(i int16) {
	b.UnsignedShort(uint16(i))
}

// Float writes a float to the buffer
func (b *Buffer) Float(f float32) {
	d := b.scratch[:4]
	i := math.Float32bits(f)
	nativeOrder.PutUint32(d, i)
	b.buf.Write(d)
}

// WriteBuffer copies the passed buffer to this buffer
func (b *Buffer) WriteBuffer(o *Buffer) {
	o.buf.WriteTo(&b.buf)
}

// Count returns the number of vertices in the buffer
func (b *Buffer) Count() int {
	return b.buf.Len() / b.elemSize
}

// Data returns a byte slice of the buffer
func (b *Buffer) Data() []byte {
	return b.buf.Bytes()
}

// ElementSize returns the size of a single vertex in the
// buffer
func (b *Buffer) ElementSize() int {
	return b.elemSize
}
