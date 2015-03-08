package render

import (
	"github.com/thinkofdeath/goquake/render/gl"
	"reflect"
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

func loadShaderAttribsUniforms(shader interface{}, program gl.Program) {
	t := reflect.TypeOf(shader).Elem()
	v := reflect.ValueOf(shader).Elem()
	l := t.NumField()

	gla := reflect.TypeOf(gl.Attribute(0))
	glu := reflect.TypeOf(gl.Uniform(0))

	for i := 0; i < l; i++ {
		f := t.Field(i)
		if f.Type == gla {
			name := f.Tag.Get("gl")
			v.Field(i).Set(reflect.ValueOf(
				program.AttributeLocation(name),
			))
		} else if f.Type == glu {
			name := f.Tag.Get("gl")
			v.Field(i).Set(reflect.ValueOf(
				program.UniformLocation(name),
			))
		}
	}
}
