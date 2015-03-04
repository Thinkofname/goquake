package bsp

import (
	"encoding/binary"
	"io"
)

const (
	sizeTextureInfo = 4*6 + 4*2 + 4*2
)

type File struct {
	lightMaps   []byte
	textures    []*texture
	textureInfo []*textureInfo
}

func ParseBSPFile(r *io.SectionReader) (bsp *File, err error) {
	bsp = &File{}

	var header bspHeader
	err = binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return
	}

	// Grab the light maps out of the file
	r.Seek(int64(header.LightMaps.Offset), 0)
	bsp.lightMaps = make([]byte, header.LightMaps.Size)
	io.ReadFull(r, bsp.lightMaps)

	// Textures
	err = bsp.parseTextures(io.NewSectionReader(r, int64(header.WallTextures.Offset), 0xFFFFFF))
	if err != nil {
		return
	}

	// Texture Info
	err = bsp.parseTextureInfo(
		io.NewSectionReader(r, int64(header.TextureInfo.Offset), 0xFFFFFF),
		int(header.TextureInfo.Size/sizeTextureInfo),
	)
	if err != nil {
		return
	}
	return
}

type bspHeader struct {
	Version        int32
	Entities       bspEntry
	Planes         bspEntry
	WallTextures   bspEntry
	Vertices       bspEntry
	VisibilityList bspEntry
	Nodes          bspEntry
	TextureInfo    bspEntry
	Faces          bspEntry
	LightMaps      bspEntry
	ClipNodes      bspEntry
	Leaves         bspEntry
	FaceList       bspEntry
	Edges          bspEntry
	Ledges         bspEntry
	Models         bspEntry
}

type bspEntry struct {
	Offset int32
	Size   int32
}

// Trims the string to the first 0 byte
func fromCString(b []byte) string {
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}