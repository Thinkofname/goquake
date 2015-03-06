package render

import (
	"github.com/thinkofdeath/goquake/render/gl"
)

const (
	gameVertexSource = `
attribute vec3 a_Position;
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

void main() {
  gl_Position = pMat * uMat * vec4(a_Position, 1.0);
  v_tex = a_tex;
  v_texInfo = a_texInfo;
  v_light = a_light;
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
    light = clamp((1.0 - (texture2D(textureLight, v_lightInfo).r)), 0.0, 63.0/64.0); // TODO: Fix the clamp
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

func compileProgram(vertex, fragment string) gl.Program {
	program := gl.CreateProgram()

	v := gl.CreateShader(gl.VertexShader)
	v.Source(vertex)
	v.Compile()

	if v.Parameter(gl.CompileStatus) == 0 {
		panic(v.InfoLog())
	}

	f := gl.CreateShader(gl.FragmentShader)
	f.Source(fragment)
	f.Compile()

	if f.Parameter(gl.CompileStatus) == 0 {
		panic(f.InfoLog())
	}

	program.AttachShader(v)
	program.AttachShader(f)
	program.Link()
	program.Use()
	return program
}
