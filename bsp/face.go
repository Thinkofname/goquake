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

	headers := make([]faceData, count)
	err := binary.Read(r, binary.LittleEndian, headers)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		header := headers[i]
		bsp.faces[i] = &face{
			plane:     bsp.planes[header.PlaneID],
			front:     header.Side == 0,
			ledges:    bsp.ledges[header.LedgeId : header.LedgeId+int32(header.LedgeNum)],
			texInfo:   bsp.textureInfo[header.TexInfoID],
			typeLight: header.TypeLight,
			baseLight: header.BaseLight,
			light:     header.Light,
			lightMap:  header.LightMap,
		}
	}
	return nil
}
