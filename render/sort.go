package render

import (
	"github.com/thinkofdeath/goquake/bsp"
)

type ti struct {
	id      int
	texture *bsp.Texture
}

type tiSorter []ti

func (t tiSorter) Len() int {
	return len(t)
}

func (t tiSorter) Less(i, j int) bool {
	if t[i].texture.Width == t[j].texture.Width {
		return t[i].texture.Height > t[j].texture.Height
	}
	return t[i].texture.Width > t[j].texture.Width
}

func (t tiSorter) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type li struct {
	id  int
	pic *bsp.Picture
}

type liSorter []li

func (l liSorter) Len() int {
	return len(l)
}

func (l liSorter) Less(i, j int) bool {
	if l[i].pic.Width == l[j].pic.Width {
		return l[i].pic.Height > l[j].pic.Height
	}
	return l[i].pic.Width > l[j].pic.Width
}

func (l liSorter) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
