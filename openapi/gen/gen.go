package gen

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Kegian/agen/openapi/parser"

	oa "github.com/swaggest/openapi-go/openapi3"
	"gopkg.in/yaml.v2"
)

func GenerateSpec(doc parser.Document) (string, error) {
	spec := oa.Spec{
		Openapi: "3.0.2",
		Info: oa.Info{
			Title:   doc.Settings.Title,
			Version: doc.Settings.Version,
		},
		Servers: []oa.Server{
			{
				URL: doc.Settings.URL,
			},
		},
		Security: []map[string][]string{
			{},
			{
				"bearerAuth": []string{},
			},
		},
	}

	for _, t := range doc.API.Tags {
		spec.Tags = append(
			spec.Tags,
			oa.Tag{
				Name:        t.Name,
				Description: nilStr(t.Description),
			},
		)
	}

	parsedPaths := map[string][]parser.Method{}
	pathsOrder := []string{}
	for _, m := range doc.API.Methods {
		if _, ok := parsedPaths[m.Path]; !ok {
			pathsOrder = append(pathsOrder, m.Path)
		}
		parsedPaths[m.Path] = append(parsedPaths[m.Path], m)
	}

	paths := oa.Paths{
		MapOfPathItemValues: map[string]oa.PathItem{},
	}
	for path, methods := range parsedPaths {
		operations := map[string]oa.Operation{}
		for _, m := range methods {
			operations[strings.ToLower(m.Method)] = GenOperation(m)
		}
		paths.MapOfPathItemValues[path] = oa.PathItem{
			MapOfOperationValues: operations,
		}
	}

	spec.Paths = paths

	schemas := map[string]oa.SchemaOrRef{}
	for _, s := range doc.Schemas {
		schemas[s.Name] = *GenSchemaOrRef(s)
	}

	spec.Components = &oa.Components{
		Schemas: &oa.ComponentsSchemas{
			MapOfSchemaOrRefValues: schemas,
		},
		SecuritySchemes: &oa.ComponentsSecuritySchemes{
			MapOfSecuritySchemeOrRefValues: map[string]oa.SecuritySchemeOrRef{
				"bearerAuth": {
					SecurityScheme: &oa.SecurityScheme{
						HTTPSecurityScheme: &oa.HTTPSecurityScheme{
							Scheme: "bearer",
						},
					},
				},
			},
		},
	}

	out, err := marshalYAML(&spec, pathsOrder)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func GenOperation(m parser.Method) oa.Operation {
	op := oa.Operation{
		ID:          nilStr(m.Name),
		Tags:        []string{m.Tag},
		Description: nilStr(m.Description),
	}

	GenOperationRequest(&op, m)
	GenOperationResponse(&op, m)

	//TODO Gen Response
	return op
}

func GenOperationRequest(op *oa.Operation, m parser.Method) {
	// Add path params
	for _, p := range m.Request.Params {
		op.Parameters = append(op.Parameters, GenParam(p))
	}
	// Add query params
	for _, p := range m.Request.Query {
		op.Parameters = append(op.Parameters, GenQuery(p))
	}

	// Add body
	if m.Request.Body != nil {
		op.RequestBody = &oa.RequestBodyOrRef{
			RequestBody: &oa.RequestBody{
				Required: nilBool(true),
				Content: map[string]oa.MediaType{
					GenContentType(m.Request.Body): {
						Schema: GenSchemaOrRef(*m.Request.Body),
					},
				},
			},
		}
	}
}

func GenContentType(s *parser.Schema) string {
	switch s.Type {
	case parser.TypeFile:
		return "application/octet-stream"
	default:
		return "application/json"
	}
}

func GenParam(p parser.Schema) oa.ParameterOrRef {
	return oa.ParameterOrRef{
		Parameter: &oa.Parameter{
			Name:        p.Name,
			In:          oa.ParameterInPath,
			Required:    nilBool(true),
			Schema:      GenSchemaOrRef(p),
			Description: nilStr(p.Description),
		},
	}
}

func GenQuery(p parser.Schema) oa.ParameterOrRef {
	return oa.ParameterOrRef{
		Parameter: &oa.Parameter{
			Name:        p.Name,
			In:          oa.ParameterInQuery,
			Required:    nilBool(!p.Optional),
			Schema:      GenSchemaOrRef(p),
			Description: nilStr(p.Description),
		},
	}
}

func GenOperationResponse(op *oa.Operation, m parser.Method) {
	resp := oa.Responses{}

	// Add default response
	if m.Response.Default != nil {
		resp.Default = GenResponse(*m.Response.Default, "")
	}

	codes := map[string]oa.ResponseOrRef{}

	// Add 200 OK response
	if m.Response.Body != nil {
		codes["200"] = *GenResponse(*m.Response.Body, "200")
	}

	// Add all error responses
	for code, body := range m.Response.Errors {
		codes[code] = *GenResponse(*body, code)
	}

	resp.MapOfResponseOrRefValues = codes
	op.Responses = resp
}

func GenResponse(s parser.Schema, code string) *oa.ResponseOrRef {
	var desc string
	switch code {
	case "":
		desc = "Default response"
	case "200":
		desc = "Successful operation"
	default:
		desc = "Response on HTTP code " + code
	}
	res := &oa.ResponseOrRef{
		Response: &oa.Response{
			Description: desc,
			Content: map[string]oa.MediaType{
				GenContentType(&s): {
					Schema: GenSchemaOrRef(s),
				},
			},
		},
	}

	if s.Type == parser.TypeFile {
		res.Response.Headers = map[string]oa.HeaderOrRef{
			"Content-Disposition": {
				Header: &oa.Header{
					Schema: &oa.SchemaOrRef{
						Schema: &oa.Schema{
							Type: nilType(parser.TypeString),
						},
					},
					Example:     nilAny("attachment; filename=\"name.csv\""),
					Description: nilStr("example: `attachment; filename=\"name.csv\"`"),
				},
			},
		}
	}

	return res
}

func GenSchemaOrRef(s parser.Schema) *oa.SchemaOrRef {
	if !s.IsArray && s.Type.IsRef() {
		return &oa.SchemaOrRef{
			SchemaReference: &oa.SchemaReference{
				Ref: "#/components/schemas/" + s.Type.Name(),
			},
		}
	}

	schema := &oa.Schema{
		Description: nilStr(s.Description),
		Example:     nilAny(s.Example),
		Type:        nilType(s.Type),
		Format:      nilFormat(s.Type),
	}

	if s.IsArray {
		t := oa.SchemaTypeArray
		schema.Type = &t
		schema.Format = nil
		items := &oa.SchemaOrRef{}
		if s.Type.IsRef() {
			items.SchemaReference = &oa.SchemaReference{
				Ref: "#/components/schemas/" + s.Type.Name(),
			}
		} else {
			items.Schema = &oa.Schema{
				Type:   nilType(s.Type),
				Format: nilFormat(s.Type),
			}
		}
		schema.Items = items
	}

	if s.Type == parser.TypeObject {
		schema.Properties = map[string]oa.SchemaOrRef{}
		for _, f := range s.Fields {
			if !f.Optional {
				schema.Required = append(schema.Required, f.Name)
			}
			prop := GenSchemaOrRef(f)
			schema.Properties[f.Name] = *prop
		}

	}

	return &oa.SchemaOrRef{Schema: schema}
}

func nilStr(s string) *string {
	if s == "" {
		return nil
	}

	tmp := s
	return &tmp
}

func nilBool(b bool) *bool {
	tmp := b
	return &tmp
}

func nilAny(s string) *any {
	if s == "" {
		return nil
	}
	var tmp any = s
	return &tmp
}

func nilType(p parser.Type) *oa.SchemaType {
	var t oa.SchemaType

	switch p {
	case parser.TypeAny:
		return nil
	case parser.TypeBool:
		t = oa.SchemaTypeBoolean
	case parser.TypeString:
		t = oa.SchemaTypeString
	case parser.TypeUUID:
		t = oa.SchemaTypeString
	case parser.TypeFile:
		t = oa.SchemaTypeString
	case parser.TypeInt32:
		t = oa.SchemaTypeInteger
	case parser.TypeInt64:
		t = oa.SchemaTypeInteger
	case parser.TypeFloat:
		t = oa.SchemaTypeNumber
	case parser.TypeDouble:
		t = oa.SchemaTypeNumber
	case parser.TypeObject:
		t = oa.SchemaTypeObject
	}

	return &t
}

func nilFormat(p parser.Type) *string {
	var f string

	switch p {
	case parser.TypeInt32:
		f = "int32"
	case parser.TypeInt64:
		f = "int64"
	case parser.TypeFloat:
		f = "float"
	case parser.TypeDouble:
		f = "double"
	case parser.TypeUUID:
		f = "uuid"
	case parser.TypeFile:
		f = "binary"
	default:
		return nil
	}

	return &f
}

func marshalYAML(s *oa.Spec, pathsOrder []string) ([]byte, error) {
	jsonData, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var v orderedMap

	err = json.Unmarshal(jsonData, &v)
	if err != nil {
		return nil, err
	}
	pathsIDX := 0
	for i, v := range v {
		if v.Key == "paths" {
			pathsIDX = i
			break
		}
	}
	paths := v[pathsIDX].Value.(map[string]any)

	orderedPaths := orderedMap{}
	for _, p := range pathsOrder {
		orderedPaths = append(
			orderedPaths,
			yaml.MapItem{
				Key:   p,
				Value: paths[p],
			},
		)
	}

	v[pathsIDX] = yaml.MapItem{
		Key:   "paths",
		Value: yaml.MapSlice(orderedPaths),
	}

	return yaml.Marshal(yaml.MapSlice(v))
}

type orderedMap []yaml.MapItem

func (om *orderedMap) UnmarshalJSON(data []byte) error {
	var mapData map[string]interface{}

	err := json.Unmarshal(data, &mapData)
	if err != nil {
		return err
	}

	keys, err := objectKeys(data)
	if err != nil {
		return err
	}

	for _, key := range keys {
		*om = append(*om, yaml.MapItem{
			Key:   key,
			Value: mapData[key],
		})
	}

	return nil
}

func objectKeys(b []byte) ([]string, error) {
	d := json.NewDecoder(bytes.NewReader(b))

	t, err := d.Token()
	if err != nil {
		return nil, err
	}

	if t != json.Delim('{') {
		return nil, errors.New("expected start of object")
	}

	var keys []string

	for {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}

		if t == json.Delim('}') {
			return keys, nil
		}

		keys = append(keys, t.(string))

		if err := skipValue(d); err != nil {
			return nil, err
		}
	}
}

var errUnterminated = errors.New("unterminated array or object")

func skipValue(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}

	switch t {
	case json.Delim('['), json.Delim('{'):
		for {
			if err := skipValue(d); err != nil {
				if errors.Is(err, errUnterminated) {
					break
				}

				return err
			}
		}
	case json.Delim(']'), json.Delim('}'):
		return errUnterminated
	}

	return nil
}
