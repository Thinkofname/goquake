package bsp

import (
	"encoding/binary"
	"io"
)

type Texture struct {
	ID            int
	Name          string
	Width, Height int
	Pictures      [4]*Picture
}

func (bsp *File) parseTextures(r *io.SectionReader) error {
	var count int32
	err := binary.Read(r, binary.LittleEndian, &count)
	if err != nil {
		return err
	}

	bsp.Textures = make([]*Texture, count)

	for i := 0; i < int(count); i++ {
		var offset int32
		err = binary.Read(r, binary.LittleEndian, &offset)
		if err != nil {
			return err
		}

		if offset == -1 {
			continue
		}

		tex, err := parseTexture(io.NewSectionReader(r, int64(offset), 0xFFFFFF))
		if err != nil {
			return err
		}
		tex.ID = i
		// Textures are referred to by index not by name
		bsp.Textures[i] = tex
	}
	return nil
}

func parseTexture(r *io.SectionReader) (*Texture, error) {
	var tex textureData
	err := binary.Read(r, binary.LittleEndian, &tex)
	if err != nil {
		return nil, err
	}

	t := &Texture{
		Name:   fromCString(tex.Name[:]),
		Width:  int(tex.Width),
		Height: int(tex.Height),
	}

	for i := uint(0); i < 4; i++ {
		t.Pictures[i] = readPicture(
			r,
			int64(tex.Offsets[i]),
			t.Width>>i,
			t.Height>>i,
		)
	}

	return t, nil
}

type Picture struct {
	Width, Height int
	Data          []byte
}

func readPicture(r *io.SectionReader, offset int64, width, height int) *Picture {
	data := make([]byte, width*height)
	io.ReadFull(r, data)
	return &Picture{
		Width:  width,
		Height: height,
		Data:   data,
	}
}

type textureData struct {
	Name    [16]byte
	Width   uint32
	Height  uint32
	Offsets [4]uint32
}
