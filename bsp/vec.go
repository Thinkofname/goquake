package bsp

import (
	"encoding/binary"
	"github.com/thinkofdeath/goquake/vmath"
	"io"
)

func (bsp *File) parseVertices(r *io.SectionReader, count int) error {
	bsp.vertices = make([]vmath.Vector3, count)
	return binary.Read(r, binary.LittleEndian, bsp.vertices)
}

type Edge struct {
	Vertex0 *vmath.Vector3
	Vertex1 *vmath.Vector3
}

type edgeData struct {
	Vertex0 uint16
	Vertex1 uint16
}

func (bsp *File) parseEdges(r *io.SectionReader, count int) error {
	bsp.Edges = make([]Edge, count)

	headers := make([]edgeData, count)
	err := binary.Read(r, binary.LittleEndian, headers)
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		bsp.Edges[i] = Edge{
			Vertex0: &bsp.vertices[headers[i].Vertex0],
			Vertex1: &bsp.vertices[headers[i].Vertex1],
		}
	}
	return nil
}
