package bsp

import (
	"encoding/binary"
	"io"
)

type vec3 struct {
	X float32
	Y float32
	Z float32
}

type vertex struct {
	vec3
}

func (bsp *File) parseVertices(r *io.SectionReader, count int) error {
	bsp.vertices = make([]vertex, count)
	return binary.Read(r, binary.LittleEndian, bsp.vertices)
}

type edge struct {
	vertex0 *vertex
	vertex1 *vertex
}

type edgeHeader struct {
	Vertex0 uint16
	Vertex1 uint16
}

func (bsp *File) parseEdges(r *io.SectionReader, count int) error {
	bsp.edges = make([]edge, count)

	headers := make([]edgeHeader, count)
	err := binary.Read(r, binary.LittleEndian, headers)
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		bsp.edges[i] = edge{
			vertex0: &bsp.vertices[headers[i].Vertex0],
			vertex1: &bsp.vertices[headers[i].Vertex1],
		}
	}
	return nil
}
