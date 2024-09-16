package parser

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type ScalarType struct {
	Type        Type
	Format      string
	Optional    bool
	IsArray     bool
	Description string
	Example     string
}

func ParseScalarType(n *yaml.Node) (ScalarType, error) {
	res := ScalarType{}

	comment := ParseComment(n.LineComment)
	res.Description = comment.Description
	res.Example = comment.Example

	val := n.Value
	if strings.HasSuffix(val, "?") {
		res.Optional = true
		val = val[:len(val)-1]
	}
	if strings.HasSuffix(val, "[]") {
		res.IsArray = true
		val = val[:len(val)-2]
	}
	if lp := strings.Index(val, "("); lp >= 0 {
		rp := strings.LastIndex(val, ")")
		if rp > lp {
			res.Format = val[lp+1 : rp]
			val = val[:lp]
		}
	}
	if strings.HasPrefix(val, "$") {
		res.Type = Type(val)
		return res, nil
	}

	var err error
	res.Type, err = GetType(val)
	if err != nil {
		return ScalarType{}, Err(n, err.Error())
	}

	return res, nil
}

type Comment struct {
	Description string
	Example     string
}

func ParseComment(comment string) Comment {
	res := Comment{}

	if !strings.HasPrefix(comment, "#") {
		return res
	}

	comment = strings.TrimSpace(comment[1:])

	if lp := strings.LastIndex(comment, "("); lp != -1 && strings.HasSuffix(comment, ")") {
		res.Example = comment[lp+1 : len(comment)-1]
		res.Description = strings.TrimSpace(comment[:lp])
	} else {
		res.Description = comment
	}

	return res
}
