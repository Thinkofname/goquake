package render

import (
	"github.com/thinkofdeath/goquake/render/gl"
)

type glTexture struct {
	Data          []byte
	Width, Height int
	Format        gl.TextureFormat
	Type          gl.Type
	Filter        gl.TextureValue
	Wrap          gl.TextureValue
}

func createTexture(t glTexture) gl.Texture {
	if t.Format == 0 {
		t.Format = gl.RGB
	}
	if t.Type == 0 {
		t.Type = gl.UnsignedByte
	}
	if t.Filter == 0 {
		t.Filter = gl.Nearest
	}
	if t.Wrap == 0 {
		t.Wrap = gl.ClampToEdge
	}

	texture := gl.CreateTexture()
	texture.Bind(gl.Texture2D)
	texture.Image2D(0, t.Format, t.Width, t.Height, t.Format, t.Type, t.Data)
	texture.Parameter(gl.TextureMagFilter, t.Filter)
	texture.Parameter(gl.TextureMinFilter, t.Filter)
	texture.Parameter(gl.TextureWrapS, t.Wrap)
	texture.Parameter(gl.TextureWrapT, t.Wrap)
	return texture
}
