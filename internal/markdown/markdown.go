package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Kegian/agen/openapi/parser"
)

func GenMarkdown(spec parser.Document) (string, error) {
	gen := &Generator{
		spec: spec,
		regs: map[string]*parser.Schema{},
	}
	return gen.Generate()
}

type Generator struct {
	bytes.Buffer
	spec parser.Document
	regs map[string]*parser.Schema
}

func (g *Generator) Add(strs ...string) {
	for _, s := range strs {
		g.WriteString(s)
	}
	g.WriteString("\n")
}

func (g *Generator) Generate() (string, error) {
	g.Add(`# <p style="text-align:center;">Endpoints</p>`)
	g.Add(`---`)
	g.Add(`---`)

	g.RegSchemas()
	for _, m := range g.spec.API.Methods {
		err := g.GenerateMethod(m)
		if err != nil {
			return "", err
		}
	}

	return g.String(), nil
}

func (g *Generator) GenerateMethod(m parser.Method) error {
	g.Add(
		`# <p style="text-align:center;"><span style="color:darkgreen;">**`,
		m.Method, " ", m.Path,
		`**</span></p>`,
	)
	g.Add()
	g.Add(`| Метод | Endpoint | Описание |`)
	g.Add(`| --- | --- | --- |`)
	g.Add(`| `, m.Method, ` | `, m.Path, ` | `, m.Description, ` |`)
	g.Add()
	g.Add(`## **Request**`)
	g.Add()
	g.Add(`### **Параметры запроса:**`)
	g.Add()

	if len(m.Request.Params) != 0 {
		g.Add(`Параметры пути`)
		g.Add()
		if err := g.GenerateParams(m.Request.Params); err != nil {
			return err
		}
		g.Add()
	}

	if len(m.Request.Query) != 0 {
		g.Add(`Параметры query`)
		g.Add()
		if err := g.GenerateParams(m.Request.Query); err != nil {
			return err
		}
		g.Add()
	}

	if m.Request.Body != nil {
		g.Add(`Параметры body`)
		g.Add()
		if err := g.GenerateBody(m.Request.Body); err != nil {
			return err
		}
		g.Add()
	}

	g.Add(`## **Response**`)
	g.Add()
	g.Add(`### **Параметры ответа:**`)
	g.Add()

	if m.Response.Body != nil {
		if err := g.GenerateBody(m.Response.Body); err != nil {
			return err
		}
	} else {
		g.Add(`empty`)
	}

	g.Add()
	g.Add(`---`)
	g.Add(`---`)
	g.Add()

	return nil
}

func (g *Generator) GenerateParams(params []parser.Schema) error {
	g.Add(`| Параметр | Тип данных | Обязательное/необязательное | Описание |`)
	g.Add(`| --- | --- | --- | --- |`)
	for _, p := range params {
		if err := g.GenerateSchema("", false, p.Description, &p); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) GenerateBody(body *parser.Schema) error {
	g.Add(`| Параметр | Тип данных | Обязательное/необязательное | Описание |`)
	g.Add(`| --- | --- | --- | --- |`)
	return g.GenerateSchema("", true, body.Description, body)
}

func (g *Generator) GenerateSchema(root string, isRef bool, desc string, s *parser.Schema) error {
	switch {
	case s.Type.IsRef():
		ref, ok := g.regs[s.Type.Name()]
		if !ok {
			return fmt.Errorf("unknown type `%s`", string(s.Type))
		}
		if !isRef {
			root += s.Name + "."
		}
		return g.GenerateSchema(root, true, s.Description, ref)
	case s.Type == parser.TypeObject:
		if !isRef {
			root += s.Name + "."
		}
		for _, f := range s.Fields {
			if err := g.GenerateSchema(root, false, "", &f); err != nil {
				return err
			}
		}
		return nil
	default:
		typ, err := getType(s.Type)
		if err != nil {
			return err
		}
		name := root
		if !isRef {
			name += s.Name
		}
		name = strings.Trim(name, ".")
		if desc == "" {
			desc = s.Description
		}
		g.Add(`| `, name, ` | `, typ, getArr(s.IsArray), ` | `, getReq(s.Optional), ` | `, desc, ` |`)
	}
	return nil
}

func (g *Generator) RegSchemas() {
	for _, s := range g.spec.Schemas {
		tmp := s
		g.regs[s.Name] = &tmp
	}
}

func getArr(arr bool) string {
	if arr {
		return "[]"
	}
	return ""
}

func getType(t parser.Type) (string, error) {
	switch t {
	case parser.TypeAny:
		return "any", nil
	case parser.TypeBool:
		return "boolean", nil
	case parser.TypeObject:
		return "object", nil
	case parser.TypeInt32:
		return "integer", nil
	case parser.TypeInt64:
		return "integer", nil
	case parser.TypeFloat:
		return "float", nil
	case parser.TypeDouble:
		return "float", nil
	case parser.TypeString:
		return "string", nil
	case parser.TypeUUID:
		return "string (uuid)", nil
	case parser.TypeFile:
		return "file", nil
	default:
		return "", fmt.Errorf("unknow type `%s`", string(t))
	}
}

func getReq(opt bool) string {
	if opt {
		return "optional"
	}
	return "required"
}
