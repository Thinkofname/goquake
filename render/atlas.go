package render

import (
	"github.com/thinkofdeath/goquake/bsp"
)

type textureAltas struct {
	width, height int
	buffer        []byte
	root          []*atlasPart
	padded        bool
	baked         bool
}

type atlasPart struct {
	x, y          int
	width, height int
	used          bool
	parts         []*atlasPart
}

type atlasTexture struct {
	x, y          int
	width, height int
}

func newAtlas(width, height int, padded bool) *textureAltas {
	a := &textureAltas{
		width:  width,
		height: height,
		padded: padded,
		buffer: make([]byte, width*height),
	}
	a.root = append(a.root, &atlasPart{
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

	w := picture.Width
	h := picture.Height
	if a.padded {
		w += 2
		h += 2
	}

	var p *atlasPart
	p, a.root = findFree(a.root, w, h)
	targetX := p.y
	targetY := p.x

	if targetX == -1 || targetY == -1 {
		panic("atlas full")
	}

	data := picture.Data
	for y := 0; y < h; y++ {
		index := (targetY+y)*a.width + targetX
		for x := 0; x < w; x++ {
			px := x
			py := y
			if a.padded {
				px--
				py--
			}
			a.buffer[index+x] = safeGetPixel(data, px, py, picture.Width, picture.Height)
		}
	}

	tx := targetX
	ty := targetY
	if a.padded {
		tx++
		ty++
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
	a.root = nil
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

func findFree(parts []*atlasPart, width, height int) (*atlasPart, []*atlasPart) {
	for _, part := range parts {
		if !part.used && part.width >= width && part.height >= height {
			if width != part.width {
				other := &atlasPart{
					x:      part.x + width,
					y:      part.y,
					width:  part.width - width,
					height: part.height,
				}
				parts = append(parts, other)
			}
			part.width = width
			if part.height-height > 0 {
				part.parts = append(part.parts, &atlasPart{
					x:      part.x,
					y:      part.y + height,
					width:  part.width,
					height: part.height - height,
				})
			}
			part.used = true
			return part, parts
		}
		var found *atlasPart
		found, part.parts = findFree(part.parts, width, height)
		if found != nil {
			return found, parts
		}
	}
	return nil, parts
}
