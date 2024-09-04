package parser

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

type Document struct {
	Settings Settings
	API      API
	Schemas  []Schema
}

type Settings struct {
	URL     string
	Version string
	Title   string
	// TODO Securities
}

type API struct {
	Tags    []Tag
	Methods []Method
}

type Tag struct {
	Name        string
	Description string
}

type Method struct {
	Method      string
	Path        string
	Description string
	Tag         string
	Name        string
	Request     Request
	Response    Response
}

type Request struct {
	Params  []Schema
	Query   []Schema
	Headers []Schema
	Body    *Schema
}

type Response struct {
	Body    *Schema
	Default *Schema
	Errors  map[string]*Schema
}

type Schema struct {
	Name        string
	Type        Type
	IsArray     bool
	Optional    bool
	Description string
	Example     string
	Embeds      []Type
	Fields      []Schema
}

type Type string

const (
	TypeAny    Type = "any"
	TypeBool   Type = "bool"
	TypeObject Type = "object"
	TypeInt32  Type = "int32"
	TypeInt64  Type = "int64"
	TypeFloat  Type = "float"
	TypeDouble Type = "double"
	TypeString Type = "string"
	TypeUUID   Type = "uuid"
	TypeFile   Type = "file"
)

func GetType(val string) (Type, error) {
	switch Type(val) {
	case "":
		return TypeAny, nil
	case TypeAny:
		return TypeAny, nil
	case TypeBool:
		return TypeBool, nil
	case TypeObject:
		return TypeObject, nil
	case TypeInt32:
		return TypeInt32, nil
	case TypeInt64:
		return TypeInt64, nil
	case TypeFloat, "float32":
		return TypeFloat, nil
	case TypeDouble, "float64":
		return TypeDouble, nil
	case TypeString:
		return TypeString, nil
	case TypeUUID:
		return TypeUUID, nil
	case TypeFile:
		return TypeFile, nil
	default:
		return Type(""), errors.New("unknown scalar type")
	}
}

func (t *Type) Name() string {
	if t.IsRef() {
		return string(*t)[1:]
	}
	return string(*t)
}

func (t *Type) IsRef() bool {
	return strings.HasPrefix(string(*t), "$")
}

func ParseDocument(data []byte) (Document, error) {
	var base yaml.Node
	var d = &base
	err := yaml.Unmarshal([]byte(data), d)
	if err != nil {
		return Document{}, Err(nil, err.Error())
	}

	if d.Kind != yaml.DocumentNode || len(d.Content) != 1 {
		return Document{}, Err(d, "should be document top level")
	}

	d = d.Content[0]

	doc := Document{}

	top, err := PairNodes(d)
	if err != nil {
		return Document{}, err
	}
	for _, p := range top {
		switch p.Left.Value {
		case "settings":
			doc.Settings, err = ParseSettings(p.Right)
			if err != nil {
				return Document{}, err
			}
		case "api":
			doc.API, err = ParseAPI(p.Right)
			if err != nil {
				return Document{}, err
			}
		case "schemas":
			doc.Schemas, err = ParseSchemas(p.Right)
			if err != nil {
				return Document{}, err
			}
		}
	}

	if err := ResolveEmbeds(&doc); err != nil {
		return Document{}, err
	}

	return doc, nil
}

func ResolveEmbeds(doc *Document) error {
	// Creating map of known types
	names := map[Type]*Schema{}
	for i := range doc.Schemas {
		names[Type("$"+doc.Schemas[i].Name)] = &doc.Schemas[i]
	}

	// Checking types circular dependence
	for _, s := range doc.Schemas {
		if err := checkCircularDependence(Type("$"+s.Name), names, &map[Type]struct{}{}); err != nil {
			return err
		}
	}

	// Resolve schemas first
	resolved := map[Type]struct{}{}
	for i := 0; i < len(doc.Schemas); i++ {
		resolveEmbeds(&doc.Schemas[i], names, &resolved)
	}

	// Resolve methods
	for i := 0; i < len(doc.API.Methods); i++ {
		resolveMethodEmbeds(&doc.API.Methods[i], names, &resolved)
	}

	return nil
}

func resolveMethodEmbeds(m *Method, names map[Type]*Schema, resolved *map[Type]struct{}) {
	// Resolve request embeds
	for i := 0; i < len(m.Request.Params); i++ {
		resolveEmbeds(&m.Request.Params[i], names, resolved)
	}
	for i := 0; i < len(m.Request.Query); i++ {
		resolveEmbeds(&m.Request.Query[i], names, resolved)
	}
	if m.Request.Body != nil {
		resolveEmbeds(m.Request.Body, names, resolved)
	}

	// Resolve response embeds
	if m.Response.Body != nil {
		resolveEmbeds(m.Response.Body, names, resolved)
	}
	if m.Response.Default != nil {
		resolveEmbeds(m.Response.Default, names, resolved)
	}
	for _, v := range m.Response.Errors {
		if v != nil {
			resolveEmbeds(v, names, resolved)
		}
	}
}
