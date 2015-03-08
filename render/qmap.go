package render

import (
	"github.com/thinkofdeath/goquake/bsp"
	"github.com/thinkofdeath/goquake/render/builder"
	"github.com/thinkofdeath/goquake/render/gl"
	"github.com/thinkofdeath/goquake/vmath"
	"math"
)

type qMap struct {
	bsp *bsp.File

	atlas      *textureAltas
	lightAtlas *textureAltas
	textures   []*atlasTexture

	mapBuffer gl.Buffer
	count     int
	stride    int
}

func newQMap(b *bsp.File) *qMap {
	m := &qMap{
		bsp:        b,
		atlas:      newAtlas(atlasSize, atlasSize, false),
		lightAtlas: newAtlas(atlasSize, atlasSize, true),
	}

	for _, texture := range b.Textures {
		tx := m.atlas.addPicture(texture.Pictures[0])
		m.textures = append(m.textures, tx)
	}
	m.atlas.bake()

	data := builder.New(
		builder.Float,         // X
		builder.Float,         // Y
		builder.Float,         // Z
		builder.UnsignedShort, // Texture X
		builder.UnsignedShort, // Texture Y
		builder.Short,         // Texture Info 1
		builder.Short,         // Texture Info 2
		builder.Short,         // Texture Info 3
		builder.Short,         // Texture Info 4
		builder.Short,         // Light Info X
		builder.Short,         // Light Info Y
		builder.UnsignedByte,  // Light
		builder.UnsignedByte,  // Light Type
	)
	m.stride = data.ElementSize()

	// Build the world
	for _, model := range b.Models {
		for _, face := range model.Faces {
			if face.TextureInfo.Texture.Name == "trigger" {
				continue
			}

			centerX := float32(0)
			centerY := float32(0)
			centerZ := float32(0)
			ec := float32(0)
			for _, l := range face.Ledges {
				if l < 0 {
					l = -l
				}
				e := b.Edges[l]
				centerX += e.Vertex0.X
				centerY += e.Vertex0.Y
				centerZ += e.Vertex0.Z
				centerX += e.Vertex1.X
				centerY += e.Vertex1.Y
				centerZ += e.Vertex1.Z
				ec += 2
			}
			centerX /= ec
			centerY /= ec
			centerZ /= ec

			light := face.BaseLight
			if face.BaseLight == 255 {
				light = 0
			}

			tOffsetX := float32(0)
			tOffsetY := float32(0)
			var lightS, lightT float32
			var lightSM, lightTM float32
			var width, height float32

			if face.TypeLight != 0xFF && face.LightMap != -1 {
				minS := float32(math.Inf(1))
				minT := float32(math.Inf(1))
				maxS := float32(math.Inf(-1))
				maxT := float32(math.Inf(-1))

				tInfo := face.TextureInfo
				for _, l := range face.Ledges {
					var vert *vmath.Vector3
					if l < 0 {
						vert = b.Edges[-l].Vertex1
					} else {
						vert = b.Edges[l].Vertex0
					}

					valS := vert.Dot(tInfo.VectorS) + tInfo.DistS
					valT := vert.Dot(tInfo.VectorT) + tInfo.DistT

					if minS > valS {
						minS = valS
					}
					if maxS < valS {
						maxS = valS
					}
					if minT > valT {
						minT = valT
					}
					if maxT < valT {
						maxT = valT
					}
				}

				lightS = float32(math.Floor(float64(minS / 16)))
				lightT = float32(math.Floor(float64(minT / 16)))
				lightSM = float32(math.Ceil(float64(maxS / 16)))
				lightTM = float32(math.Ceil(float64(maxT / 16)))

				width = (lightSM - lightS) + 1.0
				height = (lightTM - lightT) + 1.0

				pic := &bsp.Picture{
					Width:  int(width),
					Height: int(height),
					Data:   b.LightMaps[face.LightMap:],
				}
				tex := m.lightAtlas.addPicture(pic)
				tOffsetX = float32(tex.x)
				tOffsetY = float32(tex.y)
			}

			s := face.TextureInfo.VectorS
			t := face.TextureInfo.VectorT

			centerVec := vmath.Vector3{centerX, centerY, centerZ}
			centerS := centerVec.Dot(s) + face.TextureInfo.DistS
			centerT := centerVec.Dot(t) + face.TextureInfo.DistT

			tex := m.textures[face.TextureInfo.Texture.ID]

			centerTX := float32(-1.0)
			centerTY := float32(-1.0)
			if face.LightMap != -1 {
				centerTX = float32(math.Floor(float64(centerS/16))) - lightS
				centerTY = float32(math.Floor(float64(centerT/16))) - lightT
			}

			for _, l := range face.Ledges {
				var av, bv *vmath.Vector3
				if l >= 0 {
					bv = b.Edges[l].Vertex0
					av = b.Edges[l].Vertex1
				} else {
					av = b.Edges[-l].Vertex0
					bv = b.Edges[-l].Vertex1
				}

				aS := av.Dot(s) + face.TextureInfo.DistS
				aT := av.Dot(t) + face.TextureInfo.DistT

				aTX := float32(-1.0)
				aTY := float32(-1.0)
				if face.LightMap != -1 {
					aTX = float32(math.Floor(float64(aS/16))) - lightS
					aTY = float32(math.Floor(float64(aT/16))) - lightT
				}

				data.Float(model.Origin.X + av.X)
				data.Float(model.Origin.Y + av.Y)
				data.Float(model.Origin.Z + av.Z)
				data.UnsignedShort(uint16(tex.x))
				data.UnsignedShort(uint16(tex.y))
				data.Short(int16(aS))
				data.Short(int16(aT))
				data.Short(int16(face.TextureInfo.Texture.Width))
				data.Short(int16(face.TextureInfo.Texture.Height))
				data.Short(int16(tOffsetX + aTX))
				data.Short(int16(tOffsetY + aTY))
				data.UnsignedByte(light)
				data.UnsignedByte(face.TypeLight)

				bS := bv.Dot(s) + face.TextureInfo.DistS
				bT := bv.Dot(t) + face.TextureInfo.DistT

				bTX := float32(-1.0)
				bTY := float32(-1.0)
				if face.LightMap != -1 {
					bTX = float32(math.Floor(float64(bS/16))) - lightS
					bTY = float32(math.Floor(float64(bT/16))) - lightT
				}

				data.Float(model.Origin.X + bv.X)
				data.Float(model.Origin.Y + bv.Y)
				data.Float(model.Origin.Z + bv.Z)
				data.UnsignedShort(uint16(tex.x))
				data.UnsignedShort(uint16(tex.y))
				data.Short(int16(bS))
				data.Short(int16(bT))
				data.Short(int16(face.TextureInfo.Texture.Width))
				data.Short(int16(face.TextureInfo.Texture.Height))
				data.Short(int16(tOffsetX + bTX))
				data.Short(int16(tOffsetY + bTY))
				data.UnsignedByte(light)
				data.UnsignedByte(face.TypeLight)

				data.Float(model.Origin.X + centerX)
				data.Float(model.Origin.Y + centerY)
				data.Float(model.Origin.Z + centerZ)
				data.UnsignedShort(uint16(tex.x))
				data.UnsignedShort(uint16(tex.y))
				data.Short(int16(centerS))
				data.Short(int16(centerT))
				data.Short(int16(face.TextureInfo.Texture.Width))
				data.Short(int16(face.TextureInfo.Texture.Height))
				data.Short(int16(tOffsetX + centerTX))
				data.Short(int16(tOffsetY + centerTY))
				data.UnsignedByte(light)
				data.UnsignedByte(face.TypeLight)
			}
		}
	}

	m.lightAtlas.bake()

	m.mapBuffer = gl.CreateBuffer()
	m.mapBuffer.Bind(gl.ArrayBuffer)
	m.mapBuffer.Data(data.Data(), gl.StaticDraw)
	m.count = data.Count()

	texture.Bind(gl.Texture2D)
	texture.Image2D(0, gl.Luminance, atlasSize, atlasSize, gl.Luminance, gl.UnsignedByte, m.atlas.buffer)

	textureLight.Bind(gl.Texture2D)
	textureLight.Image2D(0, gl.Luminance, atlasSize, atlasSize, gl.Luminance, gl.UnsignedByte, m.lightAtlas.buffer)

	return m
}

func (m *qMap) render() {
	m.mapBuffer.Bind(gl.ArrayBuffer)

	gameShader.setupPointers(m.stride)
	gl.DrawArrays(gl.Triangles, 0, m.count)
}

func (m *qMap) cleanup() {
	m.mapBuffer.Delete()
}
