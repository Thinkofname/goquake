package bsp

import (
	"encoding/binary"
	"io"
)

type Face struct {
	plane       *plane
	front       bool
	Ledges      []int
	TextureInfo *TextureInfo
	TypeLight   uint8
	BaseLight   uint8
	light       [2]uint8
	LightMap    int32
}

type faceData struct {
	PlaneID   uint16
	Side      uint16
	LedgeId   int32
	LedgeNum  uint16
	TexInfoID uint16
	TypeLight uint8
	BaseLight uint8
	Light     [2]uint8
	LightMap  int32
}

func (bsp *File) parseFaces(r *io.SectionReader, count int) error {
	bsp.faces = make([]*Face, count)

	faces := make([]faceData, count)
	err := binary.Read(r, binary.LittleEndian, faces)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		f := faces[i]
		bsp.faces[i] = &Face{
			plane:       bsp.planes[f.PlaneID],
			front:       f.Side == 0,
			Ledges:      bsp.ledges[f.LedgeId : f.LedgeId+int32(f.LedgeNum)],
			TextureInfo: bsp.textureInfo[f.TexInfoID],
			TypeLight:   f.TypeLight,
			BaseLight:   f.BaseLight,
			light:       f.Light,
			LightMap:    f.LightMap,
		}
	}
	return nil
}
