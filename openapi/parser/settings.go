package parser

import (
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseSettings(n *yaml.Node) (Settings, error) {
	settings := Settings{
		URL:     "/api/v1",
		Version: "1.0.0",
		Title:   "Schema document",
	}

	pairs, err := PairNodes(n)
	if err != nil {
		return Settings{}, err
	}
	for _, p := range pairs {
		name := p.Left.Value
		switch name {
		case "url":
			settings.URL = strings.TrimSpace(p.Right.Value)
		case "version":
			settings.Version = strings.TrimSpace(p.Right.Value)
		case "title":
			settings.Title = strings.TrimSpace(p.Right.Value)
		case "security":
			// TODO
		}
	}

	return settings, nil
}
