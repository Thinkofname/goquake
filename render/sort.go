package render

import (
	"github.com/thinkofdeath/goquake/bsp"
)

// Used for sorting textures/lightmaps

type ti struct {
	id      int
	texture *bsp.Texture
}

type tiSorter []ti

func (t tiSorter) Len() int {
	return len(t)
}

func (t tiSorter) Less(i, j int) bool {
	return t[i].texture.Width*t[i].texture.Height > t[j].texture.Width*t[j].texture.Height
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
	return l[i].pic.Width*l[i].pic.Height > l[j].pic.Width*l[j].pic.Height
}

func (l liSorter) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
