package bsp

import (
	"encoding/binary"
	"io"
)

type texture struct {
	name          string
	width, height int
	pictures      [4]*picture
}

func (bsp *File) parseTextures(r *io.SectionReader) error {
	var count int32
	err := binary.Read(r, binary.LittleEndian, &count)
	if err != nil {
		return err
	}

	bsp.textures = make([]*texture, count)

	for i := 0; i < int(count); i++ {
		var offset int32
		err = binary.Read(r, binary.LittleEndian, &offset)
		if err != nil {
			return err
		}

		tex, err := parseTexture(io.NewSectionReader(r, int64(offset), 0xFFFFFF))
		if err != nil {
			return err
		}
		// Textures are referred to by index not by name
		bsp.textures[i] = tex
	}
	return nil
}

func parseTexture(r *io.SectionReader) (*texture, error) {
	var tex textureData
	err := binary.Read(r, binary.LittleEndian, &tex)
	if err != nil {
		return nil, err
	}

	t := &texture{
		name:   fromCString(tex.Name[:]),
		width:  int(tex.Width),
		height: int(tex.Height),
	}

	for i := uint(0); i < 4; i++ {
		t.pictures[i] = readPicture(
			r,
			int64(tex.Offsets[i]),
			t.width>>i,
			t.height>>i,
		)
	}

	return t, nil
}

type picture struct {
	width, height int
	data          []byte
}

func readPicture(r *io.SectionReader, offset int64, width, height int) *picture {
	data := make([]byte, width*height)
	io.ReadFull(r, data)
	return &picture{
		width:  width,
		height: height,
		data:   data,
	}
}

type textureData struct {
	Name    [16]byte
	Width   uint32
	Height  uint32
	Offsets [4]uint32
}
