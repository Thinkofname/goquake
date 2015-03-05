package bsp

import (
	"encoding/binary"
	"io"
)

type textureInfo struct {
	vectorS  vec3
	distS    float32
	vectorT  vec3
	distT    float32
	texture  *texture
	animated bool
}

type textureInfoHeader struct {
	VectorS   vec3
	DistS     float32
	VectorT   vec3
	DistT     float32
	TextureID uint32
	Animated  uint32
}

func (bsp *File) parseTextureInfo(r *io.SectionReader, count int) error {
	bsp.textureInfo = make([]*textureInfo, count)

	headers := make([]textureInfoHeader, count)
	err := binary.Read(r, binary.LittleEndian, headers)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		header := headers[i]
		bsp.textureInfo[i] = &textureInfo{
			vectorS:  header.VectorS,
			distS:    header.DistS,
			vectorT:  header.VectorT,
			distT:    header.DistT,
			texture:  bsp.textures[header.TextureID],
			animated: header.Animated != 0,
		}
	}
	return nil
}
