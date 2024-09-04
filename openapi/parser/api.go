package parser

import (
	"fmt"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseAPI(n *yaml.Node) (API, error) {
	tags := []Tag{}
	methods := []Method{}
	var common *Method

	pairs, err := PairNodes(n)
	if err != nil {
		return API{}, err
	}
	for _, p := range pairs {
		name := p.Left.Value
		switch name {
		case "_common":
			c, err := ParseCommon(p.Right)
			if err != nil {
				return API{}, err
			}
			common = &c
		default:
			comment := ParseComment(p.Left.LineComment)

			tag := Tag{
				Name:        p.Left.Value,
				Description: comment.Description,
			}
			tags = append(tags, tag)

			m, err := ParseMethods(p.Right, tag.Name)
			if err != nil {
				return API{}, err
			}
			methods = append(methods, m...)
		}
	}

	if common != nil {
		for i := 0; i < len(methods); i++ {
			methods[i].Request = MergeRequest(methods[i].Request, common.Request)
			methods[i].Response = MergeResponse(methods[i].Response, common.Response)
		}
	}

	if err := FillMethodsNames(methods); err != nil {
		return API{}, err
	}

	return API{
		Tags:    tags,
		Methods: methods,
	}, nil
}

func ParseMethods(n *yaml.Node, tag string) ([]Method, error) {
	var common *Method
	methods := make([]Method, 0)

	pairs, err := PairNodes(n)
	if err != nil {
		return nil, err
	}
	for _, p := range pairs {
		if p.Left.Value == "_common" {
			c, err := ParseCommon(p.Right)
			if err != nil {
				return nil, err
			}
			common = &c
			continue
		}

		m := strings.Split(p.Left.Value, " ")
		if len(m) != 2 {
			return nil, Err(p.Left, "incorrect method format")
		}
		if m[0] != "POST" && m[0] != "GET" {
			// TODO: support all methods
			return nil, Err(p.Left, "incorrect method type (POST/GET only)")
		}

		comment := ParseComment(p.Left.LineComment)

		method, err := ParseMethod(m[1], p.Right)
		if err != nil {
			return nil, err
		}
		method.Method = m[0]
		method.Description = comment.Description
		method.Tag = tag

		methods = append(methods, method)
	}

	if common != nil {
		for i := 0; i < len(methods); i++ {
			methods[i].Request = MergeRequest(methods[i].Request, common.Request)
			methods[i].Response = MergeResponse(methods[i].Response, common.Response)
		}
	}

	return methods, nil
}

func ParseCommon(n *yaml.Node) (Method, error) {
	method := Method{}
	pairs, err := PairNodes(n)
	if err != nil {
		return Method{}, err
	}
	for _, p := range pairs {
		switch p.Left.Value {
		case "request":
			method.Request, err = ParseRequest(p.Right)
			if err != nil {
				return Method{}, err
			}
		case "response":
			method.Response, err = ParseResponse(p.Right)
			if err != nil {
				return Method{}, err
			}
		default:
			return Method{}, Err(p.Left, "unknown field of a common")
		}
	}

	return method, nil
}

func ParseMethod(path string, n *yaml.Node) (Method, error) {
	method := Method{}
	pairs, err := PairNodes(n)
	if err != nil {
		return Method{}, err
	}
	for _, p := range pairs {
		switch p.Left.Value {
		case "name":
			method.Name = strings.TrimSpace(p.Right.Value)
		case "request":
			method.Request, err = ParseRequest(p.Right)
			if err != nil {
				return Method{}, err
			}
		case "response":
			method.Response, err = ParseResponse(p.Right)
			if err != nil {
				return Method{}, err
			}
		default:
			return Method{}, Err(p.Left, "unknown field of a method")
		}
	}

	path, params := ParsePathParams(path)
	method.Path = path
	method.Request.Params = MergeParams(method.Request.Params, params)

	return method, nil
}

func ParseRequest(n *yaml.Node) (Request, error) {
	req := Request{}
	pairs, err := PairNodes(n)
	if err != nil {
		return Request{}, err
	}
	for _, p := range pairs {
		switch p.Left.Value {
		case "params":
			req.Params, err = ParseParams(&p)
			if err != nil {
				return Request{}, err
			}
		case "query":
			req.Query, err = ParseQuery(&p)
			if err != nil {
				return Request{}, err
			}
		case "headers":
			req.Headers, err = ParseHeaders(&p)
			if err != nil {
				return Request{}, err
			}
		case "form":
			return Request{}, Err(p.Left, "unimplemented")
		case "body":
			req.Body, err = ParseBody(&p)
			if err != nil {
				return Request{}, err
			}
		default:
			return Request{}, Err(p.Left, "unknown field of a request")
		}
	}
	return req, nil
}

func ParsePathParams(path string) (string, []Schema) {
	re := regexp.MustCompile(`\{(\w+)(\:[\$\w]+){0,1}\}`)
	newPath := re.ReplaceAllString(path, "{$1}")

	params := map[string]Schema{}
	groups := re.FindAllSubmatch([]byte(path), -1)
	for _, g := range groups {
		pName := string(g[1])
		pType := TypeString
		if string(g[2]) != "" {
			pType = Type(string(g[2])[1:])
		}
		params[pName] = Schema{
			Name:     pName,
			Type:     pType,
			Optional: false,
		}
	}

	res := []Schema{}
	for _, s := range params {
		res = append(res, s)
	}

	return newPath, res
}

func ParseParams(p *NodePair) ([]Schema, error) {
	params, err := ParseSchema(p)
	if err != nil {
		return nil, err
	}
	if params.Type != TypeObject {
		return nil, Err(p.Right, "incorrect params format")
	}
	return params.Fields, nil
}

func ParseQuery(p *NodePair) ([]Schema, error) {
	query, err := ParseSchema(p)
	if err != nil {
		return nil, err
	}
	if query.Type != TypeObject {
		return nil, Err(p.Right, "incorrect query format")
	}
	return query.Fields, nil
}

var forbiddenHeaders = map[string]struct{}{"Accept": {}, "Authorization": {}, "Content-Type": {}}

func ParseHeaders(p *NodePair) ([]Schema, error) {
	headers, err := ParseSchema(p)
	if err != nil {
		return nil, err
	}
	if headers.Type != TypeObject {
		return nil, Err(p.Right, "incorrect headers format")
	}
	for i, f := range headers.Fields {
		headers.Fields[i].Name = textproto.CanonicalMIMEHeaderKey(f.Name)
	}
	for _, f := range headers.Fields {
		if _, ok := forbiddenHeaders[f.Name]; ok {
			return nil, Err(p.Right, "forbidden header "+f.Name)
		}
	}
	return headers.Fields, nil
}

func ParseBody(p *NodePair) (*Schema, error) {
	body, err := ParseSchema(p)
	if err != nil {
		return nil, err
	}
	return &body, nil
}

func MergeRequest(prior, minor Request) Request {
	var res Request
	res.Params = MergeParams(prior.Params, minor.Params)
	res.Query = MergeParams(prior.Query, minor.Query)
	res.Headers = MergeParams(prior.Headers, minor.Headers)
	res.Body = prior.Body // TODO: merge bodies?
	return res
}

func MergeParams(prior, minor []Schema) []Schema {
	res := make([]Schema, len(minor))
	copy(res, minor)

	idx := map[string]int{}
	for i, m := range minor {
		idx[m.Name] = i
	}

	for _, p := range prior {
		if i, ok := idx[p.Name]; ok {
			res[i] = p
			if p.Description == "" {
				res[i].Description = minor[i].Description
			}
			if p.Example == "" {
				res[i].Example = minor[i].Example
			}
		} else {
			res = append(res, p)
		}
	}

	return res
}

func ParseResponse(n *yaml.Node) (Response, error) {
	res := Response{
		Errors: map[string]*Schema{},
	}
	pairs, err := PairNodes(n)
	if err != nil {
		return Response{}, err
	}
	for _, p := range pairs {
		switch p.Left.Value {
		case "body":
			res.Body, err = ParseBody(&p)
			if err != nil {
				return Response{}, err
			}
		case "default":
			res.Default, err = ParseBody(&p)
			if err != nil {
				return Response{}, err
			}
		default:
			_, err := strconv.Atoi(p.Left.Value)
			if err != nil {
				return Response{}, Err(p.Left, "unknown field of a request")
			}
			res.Errors[p.Left.Value], err = ParseBody(&p)
			if err != nil {
				return Response{}, err
			}
		}
	}
	return res, nil
}

func MergeResponse(prior, minor Response) Response {
	var res Response
	res.Body = prior.Body

	res.Default = minor.Default
	if prior.Default != nil {
		res.Default = prior.Default
	}

	res.Errors = map[string]*Schema{}
	for k, v := range minor.Errors {
		res.Errors[k] = v
	}
	for k, v := range prior.Errors {
		res.Errors[k] = v
	}

	return res
}

func FillMethodsNames(methods []Method) error {
	nonames := make([]*Method, 0, len(methods))
	for i, m := range methods {
		if m.Name == "" {
			nonames = append(nonames, &methods[i])
		}
	}

	count := map[string]int{}
	unique := map[string]string{}
	for _, m := range methods {
		if m.Name != "" {
			count[m.Name]++
			if path, ok := unique[m.Name]; ok {
				return Err(
					nil,
					fmt.Sprintf(
						"Duplicate method name (%s) `%s` and `%s %s`",
						m.Name, path, m.Method, m.Path,
					),
				)
			}
			unique[m.Name] = m.Method + " " + m.Path
		}
	}

	paramsRe := regexp.MustCompile(`\{(\w+)(\:[\$\w]+){0,1}\}`)
	underRe := regexp.MustCompile(`_+`)
	for _, m := range nonames {
		name := paramsRe.ReplaceAllString(m.Path, "_")
		name = strings.ReplaceAll(name, "/", "_")
		name = underRe.ReplaceAllString(name, "_")
		name = strings.Trim(name, "_")
		name = SnakeToUpper(name)

		m.Name = name
		count[name]++
	}

	for _, m := range nonames {
		if count[m.Name] <= 1 {
			continue
		}
		name := SnakeToUpper(m.Method) + m.Name
		if path, ok := unique[name]; ok {
			return Err(
				nil,
				fmt.Sprintf(
					"Duplicate method name (%s) `%s` and `%s %s`",
					name, path, m.Method, m.Path,
				),
			)
		}
		m.Name = name
		unique[name] = m.Method + " " + m.Path
	}

	return nil
}
