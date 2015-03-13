// Package atlas provides a basic texture atlas for quake pictures
package atlas

import (
	"github.com/thinkofdeath/goquake/bsp"
	"math"
)

// Type is a texture atlas for storing quake
// pictures (from the bsp package). The buffer
// is public to allow easy uploading to the
// gpu.
type Type struct {
	width, height int
	Buffer        []byte
	freeSpace     []*Rect
	padding       int
	baked         bool
}

// Rect represents a location in a texture
// atlas.
type Rect struct {
	X, Y          int
	Width, Height int
}

// New creates an atlas of the specified size
// with zero padding around the textures.
func New(width, height int) *Type {
	return NewPadded(width, height, 0)
}

// NewPadded creates an atlas of the specified
// size. Textures are padded with the passed
// number of pixels around each size. This is
// useful for filtering textures without other
// textures bleeding through.
func NewPadded(width, height int, padding int) *Type {
	a := &Type{
		width:   width,
		height:  height,
		padding: padding,
		Buffer:  make([]byte, width*height),
	}
	a.freeSpace = append(a.freeSpace, &Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
	})
	return a
}

// Add adds the passed picture to the atlas and
// returns the location in the atlas. This method
// panics if the atlas has been baked or if the
// atlas is full.
func (a *Type) Add(picture *bsp.Picture) *Rect {
	if a.baked {
		panic("invalid state, atlas is baked")
	}

	// Double the padding since its for both
	// sides
	w := picture.Width + (a.padding * 2)
	h := picture.Height + (a.padding * 2)

	var target *Rect
	targetIndex := 0
	priority := math.MaxInt32
	// Search through and find the best fit for this texture
	for i, free := range a.freeSpace {
		if free.Width >= w && free.Height >= h {
			currentPriority := (free.Width - w) * (free.Height - h)
			if target == nil || currentPriority < priority {
				target = free
				priority = currentPriority
				targetIndex = i
			}

			// Perfect match, we can break early
			if priority == 0 {
				break
			}
		}
	}

	if target == nil {
		// TODO(Think) Return an error here and
		//             add a 'MustAdd' for the
		//             cases where you don't care?
		panic("atlas full")
	}

	// Copy the picture into the atlas
	CopyImage(picture.Data, a.Buffer, target.X, target.Y, w, h, a.width, a.height, a.padding)

	tx := target.X + a.padding
	ty := target.Y + a.padding

	if w == target.Width {
		target.Y += h
		target.Height -= h
		if target.Height == 0 {
			// Remove empty sections
			a.freeSpace = append(a.freeSpace[:targetIndex], a.freeSpace[targetIndex+1:]...)
		}
	} else {
		if target.Height > h {
			// Split by height
			a.freeSpace = append(
				[]*Rect{&Rect{
					target.X, target.Y + h,
					w, target.Height - h,
				}},
				a.freeSpace...,
			)
		}
		target.X += w
		target.Width -= w
	}

	return &Rect{
		X:      tx,
		Y:      ty,
		Width:  picture.Width,
		Height: picture.Height,
	}
}

// Bake causes the atlas to be uneditable allowing
// it to free up resources used in packing.
func (a *Type) Bake() {
	a.baked = true
	a.freeSpace = nil
}

// helper method that allows for out of bounds access
// to a picture. The coordinates to be changed to the
// nearest edge when out of bounds.
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

// CopyImage copies the passed picture data to the
// target buffer, accounting for padding.
func CopyImage(data, buffer []byte, targetX, targetY, w, h, width, height int, padding int) {
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
