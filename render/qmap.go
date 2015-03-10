package render

import (
	"github.com/thinkofdeath/goquake/bsp"
	"github.com/thinkofdeath/goquake/render/builder"
	"github.com/thinkofdeath/goquake/render/gl"
	"github.com/thinkofdeath/goquake/vmath"
	"math"
	"strings"
	"time"
)

type qMap struct {
	bsp *bsp.File

	atlas      *textureAltas
	lightAtlas *textureAltas
	textures   []*atlasTexture

	mapBuffer    gl.Buffer
	count        int
	stride       int
	skyBuffer    gl.Buffer
	skyBoxBuffer gl.Buffer
	skyCount     int
	skyBoxCount  int
	skyTexture   int
	skyMin       vmath.Vector3
	skyMax       vmath.Vector3
}

var (
	vertexSerializer func(*builder.Buffer, interface{})
	vertexTypes      []builder.Type
)

type mapVertex struct {
	X              float32
	Y              float32
	Z              float32
	TextureX       uint16
	TextureY       uint16
	TextureOffsetX int16
	TextureOffsetY int16
	TextureWidth   int16
	TextureHeight  int16
	LightX         int16
	LightY         int16
	Light          uint8
	LightType      uint8
}

func init() {
	vertexSerializer, vertexTypes = builder.Struct(mapVertex{})
}

func newQMap(b *bsp.File) *qMap {
	m := &qMap{
		bsp:        b,
		atlas:      newAtlas(atlasSize, atlasSize, false),
		lightAtlas: newAtlas(atlasSize, atlasSize, true),
		skyTexture: -1,
		textures:   make([]*atlasTexture, len(b.Textures)),
	}

	for i, texture := range b.Textures {
		if texture == nil {
			continue
		}
		tx := m.atlas.addPicture(texture.Pictures[0])
		m.textures[i] = tx
	}
	m.atlas.bake()

	dataNormal := builder.New(vertexTypes...)
	dataSky := builder.New(vertexTypes...)
	m.stride = dataNormal.ElementSize()

	// Build the world
	for _, model := range b.Models {
		for _, face := range model.Faces {
			if face.TextureInfo.Texture == nil || face.TextureInfo.Texture.Name == "trigger" {
				continue
			}

			var data *builder.Buffer
			var isSky bool
			if strings.HasPrefix(face.TextureInfo.Texture.Name, "sky") {
				if m.skyTexture != -1 && m.skyTexture != face.TextureInfo.Texture.ID {
					panic("too many sky textures")
				}
				m.skyTexture = face.TextureInfo.Texture.ID
				data = dataSky
				isSky = true
			} else {
				data = dataNormal
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

				// Sky things
				if isSky {
					for _, v := range []*vmath.Vector3{av, bv} {
						if model.Origin.X+v.X < m.skyMin.X {
							m.skyMin.X = model.Origin.X + v.X
						}
						if model.Origin.Y+v.Y < m.skyMin.Y {
							m.skyMin.Y = model.Origin.Y + v.Y
						}
						if model.Origin.X+v.X > m.skyMax.X {
							m.skyMax.X = model.Origin.X + v.X
						}
						if model.Origin.Y+v.Y > m.skyMax.Y {
							m.skyMax.Y = model.Origin.Y + v.Y
						}
						if model.Origin.Z+v.Z > m.skyMax.Z {
							m.skyMax.Z = model.Origin.Z + v.Z
						}
						if model.Origin.Z+v.Z < m.skyMin.Z {
							m.skyMin.Z = model.Origin.Z + v.Z
						}
					}
				}

				aS := av.Dot(s) + face.TextureInfo.DistS
				aT := av.Dot(t) + face.TextureInfo.DistT

				aTX := float32(-1.0)
				aTY := float32(-1.0)
				if face.LightMap != -1 {
					aTX = float32(math.Floor(float64(aS/16))) - lightS
					aTY = float32(math.Floor(float64(aT/16))) - lightT
				}

				vertexSerializer(data, mapVertex{
					X:              model.Origin.X + av.X,
					Y:              model.Origin.Y + av.Y,
					Z:              model.Origin.Z + av.Z,
					TextureX:       uint16(tex.x),
					TextureY:       uint16(tex.y),
					TextureOffsetX: int16(aS),
					TextureOffsetY: int16(aT),
					TextureWidth:   int16(face.TextureInfo.Texture.Width),
					TextureHeight:  int16(face.TextureInfo.Texture.Height),
					LightX:         int16(tOffsetX + aTX),
					LightY:         int16(tOffsetY + aTY),
					Light:          light,
					LightType:      face.TypeLight,
				})

				bS := bv.Dot(s) + face.TextureInfo.DistS
				bT := bv.Dot(t) + face.TextureInfo.DistT

				bTX := float32(-1.0)
				bTY := float32(-1.0)
				if face.LightMap != -1 {
					bTX = float32(math.Floor(float64(bS/16))) - lightS
					bTY = float32(math.Floor(float64(bT/16))) - lightT
				}

				vertexSerializer(data, mapVertex{
					X:              model.Origin.X + bv.X,
					Y:              model.Origin.Y + bv.Y,
					Z:              model.Origin.Z + bv.Z,
					TextureX:       uint16(tex.x),
					TextureY:       uint16(tex.y),
					TextureOffsetX: int16(bS),
					TextureOffsetY: int16(bT),
					TextureWidth:   int16(face.TextureInfo.Texture.Width),
					TextureHeight:  int16(face.TextureInfo.Texture.Height),
					LightX:         int16(tOffsetX + bTX),
					LightY:         int16(tOffsetY + bTY),
					Light:          light,
					LightType:      face.TypeLight,
				})

				vertexSerializer(data, mapVertex{
					X:              model.Origin.X + centerX,
					Y:              model.Origin.Y + centerY,
					Z:              model.Origin.Z + centerZ,
					TextureX:       uint16(tex.x),
					TextureY:       uint16(tex.y),
					TextureOffsetX: int16(centerS),
					TextureOffsetY: int16(centerT),
					TextureWidth:   int16(face.TextureInfo.Texture.Width),
					TextureHeight:  int16(face.TextureInfo.Texture.Height),
					LightX:         int16(tOffsetX + centerTX),
					LightY:         int16(tOffsetY + centerTY),
					Light:          light,
					LightType:      face.TypeLight,
				})
			}
		}
	}

	m.lightAtlas.bake()

	m.mapBuffer = gl.CreateBuffer()
	m.mapBuffer.Bind(gl.ArrayBuffer)
	m.mapBuffer.Data(dataNormal.Data(), gl.StaticDraw)
	m.count = dataNormal.Count()

	m.skyBuffer = gl.CreateBuffer()
	m.skyBuffer.Bind(gl.ArrayBuffer)
	m.skyBuffer.Data(dataSky.Data(), gl.StaticDraw)
	m.skyCount = dataSky.Count()

	m.skyMax.X += 2000
	m.skyMax.Y += 2000
	m.skyMin.X -= 2000
	m.skyMin.Y -= 2000
	skyBox := builder.New(vertexTypes...)
	m.buildSkyBox(skyBox)

	m.skyBoxBuffer = gl.CreateBuffer()
	m.skyBoxBuffer.Bind(gl.ArrayBuffer)
	m.skyBoxBuffer.Data(skyBox.Data(), gl.StaticDraw)
	m.skyBoxCount = skyBox.Count()

	texture.Bind(gl.Texture2D)
	texture.Image2D(0, gl.Luminance, atlasSize, atlasSize, gl.Luminance, gl.UnsignedByte, m.atlas.buffer)

	textureLight.Bind(gl.Texture2D)
	textureLight.Image2D(0, gl.Luminance, atlasSize, atlasSize, gl.Luminance, gl.UnsignedByte, m.lightAtlas.buffer)

	return m
}

func (m *qMap) render() {
	gl.Enable(gl.StencilTest)
	gl.ColorMask(false, false, false, false)
	gl.StencilMask(0xFF)
	gl.Clear(gl.StencilBufferBit)
	gl.StencilFunc(gl.Never, 1, 0xFF)
	gl.StencilOp(gl.Replace, gl.Keep, gl.Keep)

	gameSkyShader.bind()
	m.skyBuffer.Bind(gl.ArrayBuffer)

	gameSkyShader.TimeOffset.Float(-1)

	gameSkyShader.setupPointers(m.stride)
	gl.DrawArrays(gl.Triangles, 0, m.skyCount)
	gameSkyShader.unbind()

	gl.ColorMask(true, true, true, true)
	gl.StencilMask(0x00)
	gl.StencilFunc(gl.Equal, 1, 0xFF)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.ColorBufferBit)
	gameSkyShader.bind()
	m.skyBoxBuffer.Bind(gl.ArrayBuffer)

	t := time.Second * 30
	gameSkyShader.TimeOffset.Float(float32(time.Now().UnixNano()%int64(t*2)) / float32(t))

	gameSkyShader.setupPointers(m.stride)
	gl.DrawArrays(gl.Triangles, 0, m.skyBoxCount)
	gameSkyShader.unbind()

	gl.Disable(gl.StencilTest)
	///

	gameShader.bind()
	m.mapBuffer.Bind(gl.ArrayBuffer)

	gameShader.setupPointers(m.stride)
	gl.DrawArrays(gl.Triangles, 0, m.count)
	gameShader.unbind()
}

func (m *qMap) cleanup() {
	m.mapBuffer.Delete()
	m.skyBoxBuffer.Delete()
	m.skyBuffer.Delete()
}

func (m *qMap) buildSkyBox(b *builder.Buffer) {
	tex := m.textures[m.skyTexture]

	w := int16(tex.width / 2)

	for z := 0; z < 2; z++ {
		offset := float32(100 * z)
		vertexSerializer(b, mapVertex{
			X:             m.skyMin.X,
			Y:             m.skyMin.Y,
			Z:             m.skyMax.Z + offset,
			TextureX:      uint16(tex.x + int(w)*z),
			TextureY:      uint16(tex.y),
			TextureWidth:  w,
			TextureHeight: int16(tex.height),
			LightType:     uint8(z),
		})
		vertexSerializer(b, mapVertex{
			X:             m.skyMin.X,
			Y:             m.skyMax.Y,
			Z:             m.skyMax.Z + offset,
			TextureX:      uint16(tex.x + int(w)*z),
			TextureY:      uint16(tex.y),
			TextureWidth:  w,
			TextureHeight: int16(tex.height),
			LightType:     uint8(z),
		})
		vertexSerializer(b, mapVertex{
			X:             m.skyMax.X,
			Y:             m.skyMin.Y,
			Z:             m.skyMax.Z + offset,
			TextureX:      uint16(tex.x + int(w)*z),
			TextureY:      uint16(tex.y),
			TextureWidth:  w,
			TextureHeight: int16(tex.height),
			LightType:     uint8(z),
		})

		vertexSerializer(b, mapVertex{
			X:             m.skyMin.X,
			Y:             m.skyMax.Y,
			Z:             m.skyMax.Z + offset,
			TextureX:      uint16(tex.x + int(w)*z),
			TextureY:      uint16(tex.y),
			TextureWidth:  w,
			TextureHeight: int16(tex.height),
			LightType:     uint8(z),
		})
		vertexSerializer(b, mapVertex{
			X:             m.skyMax.X,
			Y:             m.skyMax.Y,
			Z:             m.skyMax.Z + offset,
			TextureX:      uint16(tex.x + int(w)*z),
			TextureY:      uint16(tex.y),
			TextureWidth:  w,
			TextureHeight: int16(tex.height),
			LightType:     uint8(z),
		})
		vertexSerializer(b, mapVertex{
			X:             m.skyMax.X,
			Y:             m.skyMin.Y,
			Z:             m.skyMax.Z + offset,
			TextureX:      uint16(tex.x + int(w)*z),
			TextureY:      uint16(tex.y),
			TextureWidth:  w,
			TextureHeight: int16(tex.height),
			LightType:     uint8(z),
		})
	}
}
