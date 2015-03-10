package pak

import (
	"io"
)

type pakGroup struct {
	files []File
}

func Join(files ...File) File {
	return pakGroup{files: files}
}

func (g pakGroup) Reader(name string) *io.SectionReader {
	for _, f := range g.files {
		r := f.Reader(name)
		if r != nil {
			return r
		}
	}
	return nil
}

func (g pakGroup) Close() (err error) {
	for _, f := range g.files {
		e := f.Close()
		if e != nil {
			err = e
		}
	}
	return
}
