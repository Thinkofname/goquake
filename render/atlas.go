package render

import (
	"github.com/thinkofdeath/goquake/bsp"
	"math"
)

type textureAltas struct {
	width, height int
	buffer        []byte
	freeSpace     []*atlasPart
	padding       int
	baked         bool
}

type atlasPart struct {
	x, y          int
	width, height int
}

type atlasTexture struct {
	x, y          int
	width, height int
}

func newAtlas(width, height int, padding int) *textureAltas {
	a := &textureAltas{
		width:   width,
		height:  height,
		padding: padding,
		buffer:  make([]byte, width*height),
	}
	a.freeSpace = append(a.freeSpace, &atlasPart{
		x:      0,
		y:      0,
		width:  width,
		height: height,
	})
	return a
}

func (a *textureAltas) addPicture(picture *bsp.Picture) *atlasTexture {
	if a.baked {
		panic("invalid state, atlas is baked")
	}

	w := picture.Width + (a.padding << 1)
	h := picture.Height + (a.padding << 1)

	var target *atlasPart
	targetIndex := 0
	priority := math.MaxInt32
	for i, free := range a.freeSpace {
		if free.width >= w && free.height >= h {
			currentPriority := free.width - w
			if currentPriority > free.height-h {
				currentPriority = free.height - h
			}
			if target == nil || currentPriority < priority {
				target = free
				priority = currentPriority
				targetIndex = i
			}

			if priority == 0 {
				break
			}
		}
	}

	if target == nil {
		panic("atlas full")
	}

	copyImage(picture.Data, a.buffer, target.x, target.y, w, h, a.width, a.height, a.padding)

	tx := target.x + a.padding
	ty := target.y + a.padding

	if w == target.width {
		target.y += h
		target.height -= h
		if target.height == 0 {
			a.freeSpace = append(a.freeSpace[:targetIndex], a.freeSpace[targetIndex+1:]...)
		}
	} else {
		if target.height > h {
			a.freeSpace = append(
				[]*atlasPart{&atlasPart{
					target.x, target.y + h,
					w, target.height - h,
				}},
				a.freeSpace...,
			)
		}
		target.x += w
		target.width -= w
	}

	return &atlasTexture{
		x:      tx,
		y:      ty,
		width:  picture.Width,
		height: picture.Height,
	}
}

func (a *textureAltas) bake() {
	a.baked = true
	a.freeSpace = nil
}

func safeGetPixel(data []byte, x, y, w, h int) byte {
	if x < 0 {
		x = 0
	}
	if x >= w {
		x = w - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= h {
		y = h - 1
	}
	return data[y*w+x]
}

func copyImage(data, buffer []byte, targetX, targetY, w, h, width, height int, padding int) {
	for y := 0; y < h; y++ {
		index := (targetY+y)*width + targetX
		for x := 0; x < w; x++ {
			px := x - padding
			py := y - padding
			pw := w - (padding << 1)
			ph := h - (padding << 1)
			buffer[index+x] = safeGetPixel(data, px, py, pw, ph)
		}
	}
}
