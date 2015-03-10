package builder

import (
	"bytes"
	"encoding/binary"
	"math"
	"unsafe"
)

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

type Type int

type Buffer struct {
	buf      bytes.Buffer
	elemSize int

	scratch [8]byte
}

func New(types ...Type) *Buffer {
	elemSize := 0
	for _, t := range types {
		elemSize += int(t)
	}
	b := &Buffer{
		elemSize: elemSize,
	}
	b.buf.Grow(elemSize * 100)
	return b
}

func (b *Buffer) UnsignedByte(i byte) {
	b.buf.WriteByte(i)
}

func (b *Buffer) Byte(i int8) {
	b.UnsignedByte(byte(i))
}

func (b *Buffer) UnsignedShort(i uint16) {
	d := b.scratch[:2]
	nativeOrder.PutUint16(d, i)
	b.buf.Write(d)
}

func (b *Buffer) Short(i int16) {
	b.UnsignedShort(uint16(i))
}

func (b *Buffer) Float(f float32) {
	d := b.scratch[:4]
	i := math.Float32bits(f)
	nativeOrder.PutUint32(d, i)
	b.buf.Write(d)
}

func (b *Buffer) WriteBuffer(o *Buffer) {
	o.buf.WriteTo(&b.buf)
}

func (b *Buffer) Count() int {
	return b.buf.Len() / b.elemSize
}

func (b *Buffer) Data() []byte {
	return b.buf.Bytes()
}

func (b *Buffer) ElementSize() int {
	return b.elemSize
}
