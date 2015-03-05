package bsp

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	sizeTextureInfo = 4*6 + 4*2 + 4*2
	sizeVertex      = 4 * 3
	sizeEdge        = 2 + 2
	sizePlane       = 4*3 + 4 + 4
	sizeFace        = 2 + 2 + 4 + 2 + 2 + 4 + 4
	sizeModel       = (4*3)*3 + 4*4 + 4 + 4 + 4
)

type File struct {
	lightMaps   []byte
	textures    []*texture
	textureInfo []*textureInfo
	vertices    []vertex
	edges       []edge
	ledges      []int
	planes      []*plane
	faces       []*face
	models      []*model
}

func ParseBSPFile(r *io.SectionReader) (bsp *File, err error) {
	bsp = &File{}

	var header bspHeader
	err = binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return
	}

	if header.Version != 29 {
		err = errors.New("unsupported version")
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

	// Vertices
	err = bsp.parseVertices(
		io.NewSectionReader(r, int64(header.Vertices.Offset), 0xFFFFFF),
		int(header.Vertices.Size/sizeVertex),
	)
	if err != nil {
		return
	}

	// Edges
	err = bsp.parseEdges(
		io.NewSectionReader(r, int64(header.Edges.Offset), 0xFFFFFF),
		int(header.Edges.Size/sizeEdge),
	)
	if err != nil {
		return
	}

	// Ledges
	ledges := make([]int32, header.Ledges.Size/4)
	r.Seek(int64(header.Ledges.Offset), 0)
	err = binary.Read(r, binary.LittleEndian, ledges)
	if err != nil {
		return
	}
	bsp.ledges = make([]int, len(ledges))
	for i, j := range ledges {
		bsp.ledges[i] = int(j)
	}

	// Planes
	err = bsp.parsePlanes(
		io.NewSectionReader(r, int64(header.Planes.Offset), 0xFFFFFF),
		int(header.Planes.Size/sizePlane),
	)
	if err != nil {
		return
	}

	// Planes
	err = bsp.parseFaces(
		io.NewSectionReader(r, int64(header.Faces.Offset), 0xFFFFFF),
		int(header.Faces.Size/sizeFace),
	)
	if err != nil {
		return
	}

	// Models
	err = bsp.parseModels(
		io.NewSectionReader(r, int64(header.Models.Offset), 0xFFFFFF),
		int(header.Models.Size/sizeModel),
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
