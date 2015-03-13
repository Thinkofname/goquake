package render

import (
	"github.com/thinkofdeath/goquake/bsp"
	"github.com/thinkofdeath/goquake/pak"
	"github.com/thinkofdeath/goquake/render/builder"
	"github.com/thinkofdeath/goquake/render/gl"
	"github.com/thinkofdeath/goquake/vmath"
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type qMap struct {
	bsp *bsp.File

	atlas       *textureAltas
	mipTextures [3][]byte
	lightAtlas  *textureAltas
	textures    []*atlasTexture

	mapVertexArray gl.VertexArray
	mapBuffer      gl.Buffer
	count          int
	stride         int

	skyVertexArray    gl.VertexArray
	skyBuffer         gl.Buffer
	skyBoxVertexArray gl.VertexArray
	skyBoxBuffer      gl.Buffer
	skyCount          int
	skyBoxCount       int
	skyTexture        int
	skyMin            vmath.Vector3
	skyMax            vmath.Vector3
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
		bsp:   b,
		atlas: newAtlas(atlasSize, atlasSize, 0),
		// Pad the light buffer to fix issues with smoothing
		// the texture
		lightAtlas: newAtlas(atlasSize, atlasSize, 1),
		skyTexture: -1,
		textures:   make([]*atlasTexture, len(b.Textures)),
	}

    // Allocate mipmaps
	for j := 0; j < 3; j++ {
		size := atlasSize >> uint(j+1)
		m.mipTextures[j] = make([]byte, size*size)
	}

	// Sort the textures by size for better packing
	var tList []ti
	for i, texture := range b.Textures {
		if texture == nil {
			continue
		}

		tList = append(tList, ti{i, texture})
	}
	sort.Sort(tiSorter(tList))

	// Add all textures to the atlas
	for _, t := range tList {
		tx := m.atlas.addPicture(t.texture.Pictures[0])
		m.textures[t.id] = tx

		// Mipmaps
		for j := 0; j < 3; j++ {
			size := atlasSize >> uint(j+1)
			copyImage(
				t.texture.Pictures[1+j].Data,
				m.mipTextures[j],
				tx.x>>uint(j+1),
				tx.y>>uint(j+1),
				tx.width>>uint(j+1),
				tx.height>>uint(j+1),
				size, size,
				0,
			)
		}
	}
	m.atlas.bake()

	bufferNormal := builder.New(vertexTypes...)
	bufferSky := builder.New(vertexTypes...)
	m.stride = bufferNormal.ElementSize()

	var lList []li

	// Search for light textures
	for _, model := range b.Models {
		for _, face := range model.Faces {
			if face.TextureInfo.Texture == nil || face.TextureInfo.Texture.Name == "trigger" {
				continue
			}
			if face.LightMap == -1 || face.TypeLight == 0xFF{
				continue
			}

			minS := math.Inf(1)
			minT := math.Inf(1)
			maxS := math.Inf(-1)
			maxT := math.Inf(-1)

			tInfo := face.TextureInfo
			for _, l := range face.Ledges {
				var vert *vmath.Vector3
				if l < 0 {
					vert = b.Edges[-l].Vertex1
				} else {
					vert = b.Edges[l].Vertex0
				}

				valS := float64(vert.Dot(tInfo.VectorS) + tInfo.DistS)
				valT := float64(vert.Dot(tInfo.VectorT) + tInfo.DistT)

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

			lightS := math.Floor(minS / 16)
			lightT := math.Floor(minT / 16)
			lightSM := math.Ceil(maxS / 16)
			lightTM := math.Ceil(maxT / 16)

			width := (lightSM - lightS) + 1.0
			height := (lightTM - lightT) + 1.0

			lList = append(lList, li{
				int(face.LightMap),
				&bsp.Picture{
					Width:  int(width),
					Height: int(height),
					Data:   b.LightMaps[face.LightMap:],
				},
			})
		}
	}
	// Add them to an atlas
	sort.Sort(liSorter(lList))
	lights := map[int32]*atlasTexture{}
	for _, l := range lList {
		lights[int32(l.id)] = m.lightAtlas.addPicture(l.pic)
	}


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
				data = bufferSky
				isSky = true
			} else {
				data = bufferNormal
			}

			switch face.TextureInfo.Texture.Name[0] {
			case '+', '*':
				face.BaseLight = 127
				face.TypeLight = 0xFF
			}

			centerX := 0.0
			centerY := 0.0
			centerZ := 0.0
			ec := 0.0
			for _, l := range face.Ledges {
				if l < 0 {
					l = -l
				}
				e := b.Edges[l]
				centerX += float64(e.Vertex0.X)
				centerY += float64(e.Vertex0.Y)
				centerZ += float64(e.Vertex0.Z)
				centerX += float64(e.Vertex1.X)
				centerY += float64(e.Vertex1.Y)
				centerZ += float64(e.Vertex1.Z)
				ec += 2
			}
			centerX /= ec
			centerY /= ec
			centerZ /= ec

			light := face.BaseLight
			if face.BaseLight == 255 {
				light = 0
			}

			tOffsetX := 0.0
			tOffsetY := 0.0
			var lightS, lightT float64

			if face.TypeLight != 0xFF && face.LightMap != -1 {
				minS := math.Inf(1)
				minT := math.Inf(1)
				maxS := math.Inf(-1)
				maxT := math.Inf(-1)

				tInfo := face.TextureInfo
				for _, l := range face.Ledges {
					var vert *vmath.Vector3
					if l < 0 {
						vert = b.Edges[-l].Vertex1
					} else {
						vert = b.Edges[l].Vertex0
					}

					valS := float64(vert.Dot(tInfo.VectorS) + tInfo.DistS)
					valT := float64(vert.Dot(tInfo.VectorT) + tInfo.DistT)

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

				lightS = math.Floor(minS / 16)
				lightT = math.Floor(minT / 16)

				tex := lights[face.LightMap]
				tOffsetX = float64(tex.x)
				tOffsetY = float64(tex.y)
			}

			s := face.TextureInfo.VectorS
			t := face.TextureInfo.VectorT

			centerVec := vmath.Vector3{float32(centerX), float32(centerY), float32(centerZ)}
			centerS := float64(centerVec.Dot(s) + face.TextureInfo.DistS)
			centerT := float64(centerVec.Dot(t) + face.TextureInfo.DistT)

			tex := m.textures[face.TextureInfo.Texture.ID]

			centerTX := -1.0
			centerTY := -1.0
			if face.LightMap != -1 {
				centerTX = math.Floor(centerS/16) - lightS
				centerTY = math.Floor(centerT/16) - lightT
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

				aS := float64(av.Dot(s) + face.TextureInfo.DistS)
				aT := float64(av.Dot(t) + face.TextureInfo.DistT)

				aTX := -1.0
				aTY := -1.0
				if face.LightMap != -1 {
					aTX = math.Floor(aS/16) - lightS
					aTY = math.Floor(aT/16) - lightT
				}

				vertexSerializer(data, mapVertex{
					X:              model.Origin.X + float32(av.X),
					Y:              model.Origin.Y + float32(av.Y),
					Z:              model.Origin.Z + float32(av.Z),
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

				bS := float64(bv.Dot(s) + face.TextureInfo.DistS)
				bT := float64(bv.Dot(t) + face.TextureInfo.DistT)

				bTX := -1.0
				bTY := -1.0
				if face.LightMap != -1 {
					bTX = math.Floor(bS/16) - lightS
					bTY = math.Floor(bT/16) - lightT
				}

				vertexSerializer(data, mapVertex{
					X:              model.Origin.X + float32(bv.X),
					Y:              model.Origin.Y + float32(bv.Y),
					Z:              model.Origin.Z + float32(bv.Z),
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
					X:              model.Origin.X + float32(centerX),
					Y:              model.Origin.Y + float32(centerY),
					Z:              model.Origin.Z + float32(centerZ),
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

	m.mapVertexArray = gl.CreateVertexArray()
	m.mapVertexArray.Bind()
	m.mapBuffer = gl.CreateBuffer()
	m.mapBuffer.Bind(gl.ArrayBuffer)
	m.mapBuffer.Data(bufferNormal.Data(), gl.StaticDraw)
	m.count = bufferNormal.Count()
	gameShader.setupPointers(m.stride)

	m.skyVertexArray = gl.CreateVertexArray()
	m.skyVertexArray.Bind()
	m.skyBuffer = gl.CreateBuffer()
	m.skyBuffer.Bind(gl.ArrayBuffer)
	m.skyBuffer.Data(bufferSky.Data(), gl.StaticDraw)
	m.skyCount = bufferSky.Count()
	gameSkyShader.setupPointers(m.stride)

	// Stretch the sky box slightly bigger than the level
	m.skyMax.X += 2000
	m.skyMax.Y += 2000
	m.skyMin.X -= 2000
	m.skyMin.Y -= 2000
	skyBox := builder.New(vertexTypes...)
	m.buildSkyBox(skyBox)

	m.skyBoxVertexArray = gl.CreateVertexArray()
	m.skyBoxVertexArray.Bind()
	m.skyBoxBuffer = gl.CreateBuffer()
	m.skyBoxBuffer.Bind(gl.ArrayBuffer)
	m.skyBoxBuffer.Data(skyBox.Data(), gl.StaticDraw)
	m.skyBoxCount = skyBox.Count()
	gameSkyShader.setupPointers(m.stride)

	texture.Bind(gl.Texture2D)
	texture.Image2D(0, gl.Red, atlasSize, atlasSize, gl.Red, gl.UnsignedByte, m.atlas.buffer)

	for j := 0; j < 3; j++ {
		size := atlasSize >> uint(j+1)
		texture.Image2D(1+j, gl.Red, size, size, gl.Red, gl.UnsignedByte, m.mipTextures[j])
	}

	textureLight.Bind(gl.Texture2D)
	textureLight.Image2D(0, gl.Red, atlasSize, atlasSize, gl.Red, gl.UnsignedByte, m.lightAtlas.buffer)

	return m
}

func (m *qMap) render() {
	// The level is actually rendered twice once for
	// the stencil buffer and then again for the screen
	// buffer. The reason for this is to allow for Quake's
	// sky box to work correctly. Instead of being visible
	// from any open area in the level the sky box can
	// only been see if see through a 'sky' quad and is
	// otherwise hidden. To replicate this we use the stencil
	// buffer to mark out the visible sky quads and then
	// only render the sky to that area and the rest of the
	// level to the inverted selection

	// Setup the stencil buffer
	gl.Enable(gl.StencilTest)
	gl.ColorMask(false, false, false, false)
	gl.StencilMask(0xFF)
	gl.Clear(gl.StencilBufferBit)
	gl.StencilOp(gl.Keep, gl.Keep, gl.Replace)

	// Hacky but -1 for time offset just fills a single
	// color.
	// TODO(Think) Separate shader?
	gameSkyShader.bind()
	gameSkyShader.TimeOffset.Float(-1)

	// We only want the depth information
	gl.StencilMask(0x00)
	m.mapVertexArray.Bind()
	gl.DrawArrays(gl.Triangles, 0, m.count)
	gl.StencilMask(0xFF)

	// Fill the stencil buffer with the location of the sky
	// quads
	gl.StencilFunc(gl.Always, 1, 0xFF)
	m.skyVertexArray.Bind()
	gl.DrawArrays(gl.Triangles, 0, m.skyCount)

	// Disable stencil writing and re-enable
	// color writing
	gl.ColorMask(true, true, true, true)
	gl.StencilMask(0x00)
	// Only target the sky quads
	gl.StencilFunc(gl.Equal, 1, 0xFF)
	// Remove the previous depth information ready
	// for the actual render
	gl.Clear(gl.DepthBufferBit)

	m.skyVertexArray.Bind()

	// Draw the two sky planes
	m.skyBoxVertexArray.Bind()
	t := time.Second * 30
	gameSkyShader.TimeOffset.Float(float32(time.Now().UnixNano()%int64(t*2)) / float32(t))
	gl.DrawArrays(gl.Triangles, 0, m.skyBoxCount)
	gameSkyShader.unbind()

	// Draw the rest of the level to the non-sky quads
	// areas
	gl.StencilFunc(gl.Equal, 0, 0xFF)
	gameShader.bind()
	m.mapVertexArray.Bind()
	gl.DrawArrays(gl.Triangles, 0, m.count)
	gameShader.unbind()

	gl.Disable(gl.StencilTest)
}

func (m *qMap) cleanup() {
	m.mapVertexArray.Delete()
	m.skyBoxVertexArray.Delete()
	m.skyVertexArray.Delete()
	m.mapBuffer.Delete()
	m.skyBoxBuffer.Delete()
	m.skyBuffer.Delete()
}

func (m *qMap) buildSkyBox(b *builder.Buffer) {
	if m.skyTexture == -1 {
		return
	}
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

func dumpTexture(data []byte, width, height int, p pak.File, name string) {
	cm, _ := ioutil.ReadAll(pakFile.Reader("gfx/colormap.lmp"))
	pm, _ := ioutil.ReadAll(pakFile.Reader("gfx/palette.lmp"))

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i, d := range data {
		ind := int(cm[32*64+int(d)])
		img.Pix[i*4+0] = pm[ind*3+0]
		img.Pix[i*4+1] = pm[ind*3+1]
		img.Pix[i*4+2] = pm[ind*3+2]
		img.Pix[i*4+3] = 0xff
	}

	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
}
