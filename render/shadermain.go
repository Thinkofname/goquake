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

func (m *mainShader) setupPointers(stride int) {
	m.Position.Pointer(3, gl.Float, false, stride, 0)
	m.TexturePos.Pointer(2, gl.UnsignedShort, false, stride, 4*3)
	m.TextureInfo.Pointer(4, gl.Short, false, stride, 4*3+2*2)
	m.LightInfo.Pointer(2, gl.Short, false, stride, 4*3+2*6)
	m.Light.Pointer(1, gl.UnsignedByte, false, stride, 4*3+2*8)
	m.LightType.Pointer(1, gl.UnsignedByte, false, stride, 4*3+2*8+1)
}

func (m *mainShader) unbind() {
	m.Position.Disable()
	m.Light.Disable()
	m.TexturePos.Disable()
	m.TextureInfo.Disable()
	m.LightInfo.Disable()
	m.LightType.Disable()
}

const (
	gameVertexSource = `
attribute vec3 a_position;
attribute float a_light;
attribute vec2 a_tex;
attribute vec4 a_texInfo;
attribute vec2 a_lightInfo;
attribute float a_lightType;

uniform mat4 pMat;
uniform mat4 uMat;
uniform float lightStyles[11];

varying vec2 v_tex;
varying vec4 v_texInfo;
varying float v_light;
varying vec2 v_lightInfo;
varying float v_lightType;

const float invTextureSize = 1.0 / 1024;
const float invPackSize = 1.0;

void main() {
  gl_Position = pMat * uMat * vec4(a_position, 1.0);
  v_tex = a_tex;
  v_texInfo = a_texInfo * invPackSize;
  v_light = a_light / 255.0;
  v_lightInfo = a_lightInfo * invTextureSize;
  v_lightType = 1.0;
  int type = int(a_lightType + 0.5);
  for (int i = 0; i < 11; i++) {
    if (type - 1 == i) {
      v_lightType *= 1.0 - lightStyles[i];
    }
  }

}
`
	gameFragmentSource = `
precision mediump float;

uniform sampler2D palette;
uniform sampler2D colourMap;
uniform sampler2D texture;
uniform sampler2D textureLight;

varying vec2 v_tex;
varying vec4 v_texInfo;
varying float v_light;
varying vec2 v_lightInfo;
varying float v_lightType;

const float invTextureSize = 1.0 / 1024;

vec3 lookupColour(float col, float light);

void main() {
  float light = v_light;
  if (v_lightInfo.x >= 0.0) {
    light = 1.0 - (texture2D(textureLight, v_lightInfo).r);
  }
  light *= v_lightType;
  vec2 offset = mod(v_texInfo.xy, v_texInfo.zw);
  float col = texture2D(texture, (v_tex.xy + offset) * invTextureSize).r;
  gl_FragColor = vec4(lookupColour(col, light), 1.0);
}

vec3 lookupColour(float col, float light) {
  float index = texture2D(colourMap, vec2(col, light)).r;
  index = floor(index * 255.0 + 0.5);
  float x = floor(mod(index, 16.0)) / 16.0;
  float y = floor(index / 16.0) / 16.0;
  return texture2D(palette, vec2(x, y)).rgb;
}
`
)
