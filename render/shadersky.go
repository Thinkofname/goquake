package render

import (
	"github.com/thinkofdeath/goquake/render/gl"
)

type skyShader struct {
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
	TimeOffset        gl.Uniform   `gl:"timeOffset"`
}

func initSkyShader() *skyShader {
	m := &skyShader{}
	m.program = compileProgram(skyVertexSource, skyFragmentSource)

	loadShaderAttribsUniforms(m, m.program)
	return m
}

func (m *skyShader) bind() {
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
}

func (m *skyShader) setupPointers(stride int) {
	m.Position.Enable()
	m.Light.Enable()
	m.TexturePos.Enable()
	m.TextureInfo.Enable()
	m.LightInfo.Enable()
	m.LightType.Enable()

	m.Position.Pointer(3, gl.Float, false, stride, 0)
	m.TexturePos.Pointer(2, gl.UnsignedShort, false, stride, 4*3)
	m.TextureInfo.Pointer(4, gl.Short, false, stride, 4*3+2*2)
	m.LightInfo.Pointer(2, gl.Short, false, stride, 4*3+2*6)
	m.Light.Pointer(1, gl.UnsignedByte, false, stride, 4*3+2*8)
	m.LightType.Pointer(1, gl.UnsignedByte, false, stride, 4*3+2*8+1)
}

func (m *skyShader) unbind() {
}

const (
	skyVertexSource = `
#version 130
in vec3 a_position;
in float a_light;
in vec2 a_tex;
in vec4 a_texInfo;
in vec2 a_lightInfo;
in float a_lightType;

uniform mat4 pMat;
uniform mat4 uMat;
uniform float lightStyles[11];

out vec2 v_tex;
out vec4 v_texInfo;
out float v_light;
out vec2 v_lightInfo;
out vec2 v_pos;
out float v_lightType;

const float invTextureSize = 1.0 / 1024.0;
const float invPackSize = 1.0;

void main() {
  gl_Position = pMat * uMat * vec4(a_position, 1.0);
  v_tex = a_tex;
  v_texInfo = a_texInfo * invPackSize;
  v_light = a_light / 255.0;
  v_lightInfo = a_lightInfo * invTextureSize;
  v_pos = a_position.xy / 4096.0;
  v_lightType = a_lightType;
}
`
	skyFragmentSource = `
#version 130
precision mediump float;

uniform sampler2D palette;
uniform sampler2D colourMap;
uniform sampler2D texture;
uniform sampler2D textureLight;
uniform float timeOffset;
uniform mat4 pMat;
uniform mat4 uMat;

in vec2 v_pos;
in vec2 v_tex;
in vec4 v_texInfo;
in float v_light;
in vec2 v_lightInfo;
in float v_lightType;

out vec4 fragColor;

const float invTextureSize = 1.0 / 1024.0;

vec3 lookupColour(float col, float light);

void main() {
  if (timeOffset < 0.0) {
    fragColor = vec4(1.0, 0.0, 0.0, 1.0);
    return;
  }
  float light = 0.5;
  vec2 offset = mod(v_pos * 1024.0 + timeOffset * v_texInfo.z * (2.0 - v_lightType), v_texInfo.zw);
  float col = textureLod(texture, (v_tex.xy + offset) * invTextureSize, 4.0 - gl_FragCoord.w * 3000.0).r;
  fragColor = vec4(lookupColour(col, light), 1.0);
}

vec3 lookupColour(float col, float light) {
  float index = texture2D(colourMap, vec2(col, light)).r;
  index = floor(index * 255.0 + 0.5);
  if (index < 1.0) discard;
  float x = floor(mod(index, 16.0)) / 16.0;
  float y = floor(index / 16.0) / 16.0;
  return texture2D(palette, vec2(x, y)).rgb;
}
`
)
