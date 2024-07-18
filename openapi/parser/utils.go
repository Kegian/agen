package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type NodePair struct {
	Left  *yaml.Node
	Right *yaml.Node
}

func PairNodes(n *yaml.Node) ([]NodePair, error) {
	if n.Kind != yaml.MappingNode {
		return nil, Err(n, "should be map element (key: value)")
	}

	nodes := n.Content
	if len(nodes)%2 != 0 {
		return nil, Err(n, "???") // TODO
	}
	pairs := make([]NodePair, len(nodes)/2)
	for i := 0; i < len(nodes)/2; i++ {
		pairs[i].Left = nodes[i*2]
		pairs[i].Right = nodes[i*2+1]
	}
	return pairs, nil
}

func Err(n *yaml.Node, msg string) error {
	if n == nil {
		return fmt.Errorf("%v", msg)
	}
	return fmt.Errorf("%v (Line: %v, Column: %v)", msg, n.Line, n.Column)
}

func PrintPair(pair *NodePair) {
	PrintNode("(left)", pair.Left)
	PrintNode("(right)", pair.Right)
}

func PrintNode(tab string, n *yaml.Node) {
	fmt.Printf("%s %s %s(%d) %s\n", tab, n.Value, NodeKind(n.Kind), len(n.Content), NodeComment(n))
	for _, c := range n.Content {
		PrintNode(tab+"- ", c)
	}
}

func NodeComment(n *yaml.Node) string {
	s := fmt.Sprintf("[%s|%s|%s]", n.HeadComment, n.LineComment, n.FootComment)
	if s == "[||]" {
		return ""
	}
	return s
}

func NodeKind(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "DocumentNode"
	case yaml.SequenceNode:
		return "SequenceNode"
	case yaml.MappingNode:
		return "MappingNode"
	case yaml.ScalarNode:
		return "ScalarNode"
	case yaml.AliasNode:
		return "AliasNode"
	default:
		return "Unknown"
	}
}

func PrettyPrint(data interface{}) error {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(val))
	return nil
}

func PrintSchema(tab string, s Schema) {
	array := "scalar"
	if s.IsArray {
		array = "array"
	}
	optional := "required"
	if s.Optional {
		optional = "optional"
	}
	fmt.Printf("%s %s: %s (%s, %s)\n", tab, s.Name, s.Type, array, optional)
	if s.Description != "" {
		fmt.Println(tab, "   Description:", s.Description)
	}
	if s.Example != "" {
		fmt.Println(tab, "   Example:", s.Example)
	}
	for _, f := range s.Fields {
		PrintSchema(tab+" - ", f)
	}
}

func SnakeToUpper(s string) string {
	fields := strings.Split(s, "_")
	res := ""
	for _, f := range fields {
		res += cases.Title(language.English).String(f)
	}
	return res
}
