package bsp

import (
	"encoding/binary"
	"io"
)

type plane struct {
	normal vec3
	dist   float32
	t      int
}

type planeData struct {
	Normal vec3
	Dist   float32
	Type   int32
}

func (bsp *File) parsePlanes(r *io.SectionReader, count int) error {
	bsp.planes = make([]*plane, count)

	headers := make([]planeData, count)
	err := binary.Read(r, binary.LittleEndian, headers)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		header := headers[i]
		bsp.planes[i] = &plane{
			normal: header.Normal,
			dist:   header.Dist,
			t:      int(header.Type),
		}
	}
	return nil
}
