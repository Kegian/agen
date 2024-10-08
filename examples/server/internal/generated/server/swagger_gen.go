// Code generated by agen, DO NOT EDIT.

package server

import (
	_ "embed"
	"net/http"

	"github.com/flowchartsman/swaggerui"
)

//go:embed openapi.yml
var SwaggerSpec []byte

func SwaggerHandler(prefix string) http.Handler {
	return http.StripPrefix(prefix, swaggerui.Handler(SwaggerSpec))
}
