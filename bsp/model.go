package bsp

import (
	"encoding/binary"
	"github.com/thinkofdeath/goquake/vmath"
	"io"
)

type Model struct {
	bound  boundingBox
	Origin vmath.Vector3
	Faces  []*Face
}

type modelData struct {
	Bound       boundingBox
	Origin      vmath.Vector3
	NodeID      [4]int32
	NumberLeafs int32
	FaceID      int32
	FaceNum     int32
}

type boundingBox struct {
	Min vmath.Vector3
	Max vmath.Vector3
}

func (bsp *File) parseModels(r *io.SectionReader, count int) error {
	bsp.Models = make([]*Model, count)

	models := make([]modelData, count)
	err := binary.Read(r, binary.LittleEndian, models)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		m := models[i]
		bsp.Models[i] = &Model{
			bound:  m.Bound,
			Origin: m.Origin,
			Faces:  bsp.faces[m.FaceID : m.FaceID+m.FaceNum],
		}
	}
	return nil
}
