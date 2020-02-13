//go:generate go run ../cmd/builtingen ../language/reference.hlb reference.go

package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type BuiltinData struct {
	Command       string
	Documentation *Documentation
}

func GenerateBuiltins(r io.Reader) ([]byte, error) {
	doc, err := GenerateDocumentation(r)
	if err != nil {
		return nil, err
	}

	data := BuiltinData{
		Command:       fmt.Sprintf("builtingen %s", strings.Join(os.Args[1:], " ")),
		Documentation: doc,
	}

	var buf bytes.Buffer
	err = referenceTmpl.Execute(&buf, &data)
	if err != nil {
		return nil, err
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		src = buf.Bytes()
	}

	return src, nil
}

var tmplFunctions = template.FuncMap{
	"unescape": func(s string) template.HTML {
		s = strconv.Quote(s)
		return template.HTML(s[1 : len(s)-1])
	},
}

type Lookup struct {
	BuiltinsByType map[string]BuiltinsLookup
}

type BuiltinsLookup struct {
	FuncByName map[string]*Func
}

var referenceTmpl = template.Must(template.New("reference").Funcs(tmplFunctions).Parse(`
// Code generated by {{.Command}}; DO NOT EDIT.

package gen

{{define "func"}}{
	Doc: "{{unescape .Doc}}",
	Type: "{{.Type}}",
	Method: {{.Method}},
	Name: "{{.Name}}",
	{{if .Params}}Params: []Field{
		{{range $i, $param := .Params}}{
			Doc: "{{unescape $param.Doc}}",
			Variadic: {{$param.Variadic}},
			Type: "{{$param.Type}}",
			Name: "{{$param.Name}}",
		},
		{{end}}
	},{{end}}
	{{if .Options}}Options: []*Func{
		{{range $i, $func := .Options}}{{template "func" $func}}
		{{end}}
	},{{end}}
},{{end}}

var (
	Reference = Lookup{
		BuiltinsByType: map[string]BuiltinsLookup{
			{{range $i, $builtin := .Documentation.Builtins}}"{{$builtin.Type}}": BuiltinsLookup{
				FuncByName: map[string]*Func{
					{{range $i, $func := $builtin.Funcs}}"{{$func.Name}}": &Func{{template "func" $func}}
					{{end}}
					{{range $i, $func := $builtin.Methods}}"{{$func.Name}}": &Func{{template "func" $func}}
					{{end}}
				},
			},
			{{end}}
		},
	}
)
`))
