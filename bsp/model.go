package bsp

import (
	"encoding/binary"
	"io"
)

type model struct {
	bound  boundingBox
	origin vec3
	faces  []*face
}

type modelData struct {
	Bound       boundingBox
	Origin      vec3
	NodeID      [4]int32
	NumberLeafs int32
	FaceID      int32
	FaceNum     int32
}

type boundingBox struct {
	Min vec3
	Max vec3
}

func (bsp *File) parseModels(r *io.SectionReader, count int) error {
	bsp.models = make([]*model, count)

	models := make([]modelData, count)
	err := binary.Read(r, binary.LittleEndian, models)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		m := models[i]
		bsp.models[i] = &model{
			bound:  m.Bound,
			origin: m.Origin,
			faces:  bsp.faces[m.FaceID : m.FaceID+m.FaceNum],
		}
	}
	return nil
}
