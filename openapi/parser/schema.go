package parser

import (
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseSchemas(n *yaml.Node) ([]Schema, error) {
	schemas := []Schema{}

	pairs, err := PairNodes(n)
	if err != nil {
		return nil, err
	}
	for _, p := range pairs {
		s, err := ParseSchema(&p)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, s)
	}

	return schemas, nil
}

func ParseSchema(p *NodePair) (Schema, error) {
	schema := Schema{}

	if p.Left.Kind != yaml.ScalarNode {
		return Schema{}, Err(p.Left, "non scalar value")
	}

	if p.Left.Value == "" {
		return Schema{}, Err(p.Left, "empty field name")
	}
	schema.Name, schema.Embeds = ParseSchemaDefinition(p.Left.Value)

	comment := ParseComment(p.Left.LineComment)
	schema.Description = comment.Description
	schema.Example = comment.Example

	switch p.Right.Kind {
	case yaml.ScalarNode:
		scalar, err := ParseScalarType(p.Right)
		if err != nil {
			return Schema{}, err
		}
		if len(schema.Embeds) != 0 && scalar.Type != TypeObject {
			return Schema{}, Err(p.Right, "type with embeds should be object")
		}
		if scalar.Description != "" {
			schema.Description = scalar.Description
		}
		if scalar.Example != "" {
			schema.Example = scalar.Example
		}
		schema.Type = scalar.Type
		schema.Optional = scalar.Optional
		schema.IsArray = scalar.IsArray

	case yaml.MappingNode:
		schema.Type = TypeObject

		pairs, err := PairNodes(p.Right)
		if err != nil {
			return Schema{}, err
		}
		fields := []Schema{}
		for _, f := range pairs {
			s, err := ParseSchema(&f)
			if err != nil {
				return Schema{}, err
			}
			fields = append(fields, s)
		}

		schema.Fields = fields

	default:
		return Schema{}, Err(p.Right, "unknown type")
	}

	return schema, nil
}

func ParseSchemaDefinition(val string) (string, []Type) {
	embStart := strings.Index(val, "<")
	embEnd := strings.LastIndex(val, ">")
	if embStart == -1 || embEnd == -1 {
		return val, nil
	}

	typ := val[0:embStart]
	val = val[embStart+1 : embEnd]
	val = strings.ReplaceAll(val, " ", "")
	embs := []Type{}
	for _, e := range strings.Split(val, ",") {
		embs = append(embs, Type(e))
	}

	return typ, embs
}

func resolveEmbeds(s *Schema, names map[Type]*Schema, resolved *map[Type]struct{}) {
	for _, e := range s.Embeds {
		if _, ok := (*resolved)[e]; !ok {
			resolveEmbeds(names[e], names, resolved)
		}
	}
	for i := len(s.Embeds) - 1; i >= 0; i-- {
		s.Fields = MergeFields(s.Fields, names[s.Embeds[i]].Fields)
	}

	for i := 0; i < len(s.Fields); i++ {
		resolveEmbeds(&s.Fields[i], names, resolved)
	}

	(*resolved)[Type("$"+s.Name)] = struct{}{}
}

func MergeFields(prior, minor []Schema) []Schema {
	res := make([]Schema, len(minor))
	copy(res, minor)

	idx := map[string]int{}
	for i, m := range minor {
		idx[m.Name] = i
	}

	for _, p := range prior {
		if i, ok := idx[p.Name]; ok {
			res[i] = p
		} else {
			res = append(res, p)
		}
	}

	return res
}

func checkCircularDependence(name Type, names map[Type]*Schema, check *map[Type]struct{}) error {
	if _, ok := names[name]; !ok {
		return Err(nil, "schema `"+string(name)+"` is not found")
	}
	if _, ok := (*check)[name]; ok {
		return Err(nil, "circular dependence for `"+string(name)+"` found")
	}
	(*check)[name] = struct{}{}
	embeds := findAllEmbeds(names[name])
	for _, e := range embeds {
		if err := checkCircularDependence(e, names, check); err != nil {
			return err
		}
	}
	return nil
}

func findAllEmbeds(s *Schema) []Type {
	res := make([]Type, len(s.Embeds))
	copy(res, s.Embeds)
	for _, f := range s.Fields {
		res = append(res, findAllEmbeds(&f)...)
	}
	m := map[Type]struct{}{}
	for _, r := range res {
		m[r] = struct{}{}
	}
	unique := []Type{}
	for k := range m {
		unique = append(unique, k)
	}
	return unique
}
