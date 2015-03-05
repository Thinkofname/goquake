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

type textureInfoData struct {
	VectorS   vec3
	DistS     float32
	VectorT   vec3
	DistT     float32
	TextureID uint32
	Animated  uint32
}

func (bsp *File) parseTextureInfo(r *io.SectionReader, count int) error {
	bsp.textureInfo = make([]*textureInfo, count)

	textures := make([]textureInfoData, count)
	err := binary.Read(r, binary.LittleEndian, textures)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		t := textures[i]
		bsp.textureInfo[i] = &textureInfo{
			vectorS:  t.VectorS,
			distS:    t.DistS,
			vectorT:  t.VectorT,
			distT:    t.DistT,
			texture:  bsp.textures[t.TextureID],
			animated: t.Animated != 0,
		}
	}
	return nil
}
