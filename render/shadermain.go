package render

import (
	"github.com/thinkofdeath/goquake/render/gl"
)

type mainShader struct {
	program gl.Program

	Position          gl.Attribute `gl:"a_position"`
	Light             gl.Attribute `gl:"a_light"`
	TexturePos        gl.Attribute `gl:"a_tex"`
	TextureInfo       gl.Attribute `gl:"a_texInfo"`
	LightInfo         gl.Attribute `gl:"a_lightInfo"`
	LightType         gl.Attribute `gl:"a_lightType"`
	PerspectiveMatrix gl.Uniform   `gl:"pMat"`
	CameraMatrix      gl.Uniform   `gl:"uMat"`
	ColourMap         gl.Uniform   `gl:"colourMap"`
	Palette           gl.Uniform   `gl:"palette"`
	Texture           gl.Uniform   `gl:"texture"`
	TextureLight      gl.Uniform   `gl:"textureLight"`
	LightStyles       gl.Uniform   `gl:"lightStyles"`
}

func initMainShader() *mainShader {
	m := &mainShader{}
	m.program = compileProgram(gameVertexSource, gameFragmentSource)

	loadShaderAttribsUniforms(m, m.program)
	return m
}

func (m *mainShader) bind() {
	m.program.Use()
	m.PerspectiveMatrix.Matrix4(false, perspectiveMatrix)
	m.CameraMatrix.Matrix4(false, cameraMatrix)

	// Bind textures

	gl.ActiveTexture(0)
	palette.Bind(gl.Texture2D)
	m.Palette.Int(0)

	gl.ActiveTexture(1)
	colourMap.Bind(gl.Texture2D)
	m.ColourMap.Int(1)

	gl.ActiveTexture(2)
	texture.Bind(gl.Texture2D)
	m.Texture.Int(2)

	gl.ActiveTexture(3)
	textureLight.Bind(gl.Texture2D)
	m.TextureLight.Int(3)

	m.Position.Enable()
	m.Light.Enable()
	m.TexturePos.Enable()
	m.TextureInfo.Enable()
	m.LightInfo.Enable()
	m.LightType.Enable()
}

func (m *mainShader) setupPointers() {
	m.Position.Pointer(3, gl.Float, false, floatsPerVertex*4, 0)
	m.Light.Pointer(1, gl.Float, false, floatsPerVertex*4, 4*3)
	m.TexturePos.Pointer(2, gl.Float, false, floatsPerVertex*4, 4*4)
	m.TextureInfo.Pointer(4, gl.Float, false, floatsPerVertex*4, 4*6)
	m.LightInfo.Pointer(2, gl.Float, false, floatsPerVertex*4, 4*10)
	m.LightType.Pointer(1, gl.Float, false, floatsPerVertex*4, 4*12)
}

func (m *mainShader) unbind() {
	m.Position.Disable()
	m.Light.Disable()
	m.TexturePos.Disable()
	m.TextureInfo.Disable()
	m.LightInfo.Disable()
	m.LightType.Disable()
}
