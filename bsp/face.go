package bsp

import (
	"encoding/binary"
	"io"
)

type face struct {
	plane     *plane
	front     bool
	ledges    []int
	texInfo   *textureInfo
	typeLight uint8
	baseLight uint8
	light     [2]uint8
	lightMap  int32
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
	bsp.faces = make([]*face, count)

	faces := make([]faceData, count)
	err := binary.Read(r, binary.LittleEndian, faces)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		f := faces[i]
		bsp.faces[i] = &face{
			plane:     bsp.planes[f.PlaneID],
			front:     f.Side == 0,
			ledges:    bsp.ledges[f.LedgeId : f.LedgeId+int32(f.LedgeNum)],
			texInfo:   bsp.textureInfo[f.TexInfoID],
			typeLight: f.TypeLight,
			baseLight: f.BaseLight,
			light:     f.Light,
			lightMap:  f.LightMap,
		}
	}
	return nil
}
