package bsp

import (
	"encoding/binary"
	"github.com/thinkofdeath/goquake/vmath"
	"io"
)

type TextureInfo struct {
	VectorS  vmath.Vector3
	DistS    float32
	VectorT  vmath.Vector3
	DistT    float32
	Texture  *Texture
	Animated bool
}

type textureInfoData struct {
	VectorS   vmath.Vector3
	DistS     float32
	VectorT   vmath.Vector3
	DistT     float32
	TextureID uint32
	Animated  uint32
}

func (bsp *File) parseTextureInfo(r *io.SectionReader, count int) error {
	bsp.textureInfo = make([]*TextureInfo, count)

	textures := make([]textureInfoData, count)
	err := binary.Read(r, binary.LittleEndian, textures)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		t := textures[i]
		bsp.textureInfo[i] = &TextureInfo{
			VectorS:  t.VectorS,
			DistS:    t.DistS,
			VectorT:  t.VectorT,
			DistT:    t.DistT,
			Texture:  bsp.Textures[t.TextureID],
			Animated: t.Animated != 0,
		}
	}
	return nil
}
