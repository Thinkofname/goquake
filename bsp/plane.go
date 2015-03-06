package bsp

import (
	"encoding/binary"
	"github.com/thinkofdeath/goquake/vmath"
	"io"
)

type plane struct {
	normal vmath.Vector3
	dist   float32
	t      int
}

type planeData struct {
	Normal vmath.Vector3
	Dist   float32
	Type   int32
}

func (bsp *File) parsePlanes(r *io.SectionReader, count int) error {
	bsp.planes = make([]*plane, count)

	planes := make([]planeData, count)
	err := binary.Read(r, binary.LittleEndian, planes)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		p := planes[i]
		bsp.planes[i] = &plane{
			normal: p.Normal,
			dist:   p.Dist,
			t:      int(p.Type),
		}
	}
	return nil
}
