// Package pak provides methods to read PAK files
package pak

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	pakMagic  = "PACK"
	entrySize = 0x40
)

var (
	// ErrInvalid is returned when the PAK file is invalid
	ErrInvalid = errors.New("Invalid PAK file")
)

// Type contains information about every entry in the PAK
// file.
type File struct {
	r     readable
	files map[string]pakEntry
}

// FromFile creates a pak.Type from a file with the given name
func FromFile(name string) (t *File, err error) {
	file, err := os.Open(name)
	if err != nil {
		return
	}
	t, err = fromReadable(file)
	return
}

type readable interface {
	io.ReadCloser
	io.ReaderAt
}

func fromReadable(r readable) (t *File, err error) {
	var header header
	err = binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return
	}

	// Make sure the file's magic is correct so that
	// we know that we are working with a PAK file
	if string(header.Magic[:]) != pakMagic {
		err = ErrInvalid
		return
	}

	t = &File{
		r:     r,
		files: make(map[string]pakEntry),
	}

	entries := make([]entry, header.DirSize/entrySize)

	err = binary.Read(
		io.NewSectionReader(r, int64(header.DirOffset), int64(header.DirSize)),
		binary.LittleEndian,
		entries,
	)
	if err != nil {
		return
	}

	for _, e := range entries {
		name := strings.ToLower(fromCString(e.FileName[:]))
		t.files[name] = pakEntry{int64(e.Offset), int64(e.Size)}
	}
	return
}

// Reader returns a section reader for the entry with
// the given name, returns nil if the entry doesn't exist
// in this PAK file. Readers returned from this will be
// closed after Type.Close is called and should not be
// used after this point.
func (t *File) Reader(name string) *io.SectionReader {
	e, ok := t.files[strings.ToLower(name)]
	if !ok {
		return nil
	}
	return io.NewSectionReader(t.r, e.Offset, e.Size)
}

// Close closes the reader used for the PAK file
func (t *File) Close() error {
	return t.r.Close()
}

type pakEntry struct {
	Offset int64
	Size   int64
}

// Parsing helpers

// Trims the string to the first 0 byte
func fromCString(b []byte) string {
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

type header struct {
	Magic     [4]byte
	DirOffset int32
	DirSize   int32
}

type entry struct {
	FileName [0x38]byte
	Offset   int32
	Size     int32
}
