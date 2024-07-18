package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/Kegian/agen/cmd/agen/web/static"
	"github.com/Kegian/agen/internal/markdown"
	"github.com/Kegian/agen/openapi/gen"
	"github.com/Kegian/agen/openapi/parser"

	"github.com/flowchartsman/swaggerui"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var (
	addr    string
	verbose bool

	initialPath string
)

func init() {
	WebCmd.Flags().StringVarP(&addr, "addr", "a", "localhost:8777", `Address for the server`)
	WebCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, `Verbose print parsed schema`)
}

var WebCmd = &cobra.Command{
	Use:   "web <input_path>",
	Short: "Open web server with generated swagger",
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) > 0 {
			initialPath = args[0]
			if _, err := os.Stat(initialPath); errors.Is(err, os.ErrNotExist) {
				err = os.WriteFile(initialPath, []byte(templateFile), 0644)
				if err != nil {
					log.Fatal(err.Error())
				}
			}
		}

		r := mux.NewRouter()
		r.PathPrefix("/swagger/{id}/").HandlerFunc(SwaggerHandler)
		r.Handle("/file", http.HandlerFunc(ArgFileHandler))
		r.Handle("/save", http.HandlerFunc(SaveHandler))
		r.Handle("/generate", http.HandlerFunc(GenerateHandler))
		r.PathPrefix("/").Handler(http.FileServer(http.FS(static.Static)))

		fmt.Println("Web server started at http://" + addr + "/")
		if err := http.ListenAndServe(addr, r); err != nil {
			log.Fatal(err.Error())
		}

		return nil
	},
}

var swaggers = NewSwaggerBuff(50)

func SwaggerHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	swg, ok := swaggers.Find(id)
	if !ok {
		http.Error(w, "Swagger outdated, regenerate", http.StatusBadRequest)
		return
	}
	handler := http.StripPrefix("/swagger/"+id, swaggerui.Handler([]byte(swg.Spec)))
	handler.ServeHTTP(w, r)
}

type ArgFile struct {
	Text string `json:"text"`
	Path string `json:"path"`
}

func ArgFileHandler(w http.ResponseWriter, _ *http.Request) {
	var initialText = templateFile
	if initialPath != "" {
		data, err := os.ReadFile(initialPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		initialText = string(data)
	}
	DoResponse(w, &ArgFile{Text: initialText, Path: initialPath})
}

type SaveReq struct {
	Text string `json:"text"`
}

func SaveHandler(w http.ResponseWriter, r *http.Request) {
	if initialPath == "" {
		http.Error(w, "no initial file", http.StatusBadRequest)
		return
	}

	var req SaveReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = os.WriteFile(initialPath, []byte(req.Text), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	DoResponse(w, &ArgFile{Text: req.Text, Path: initialPath})
}

type GenerateReq struct {
	Text string `json:"text"`
}

type GenerateRes struct {
	OpenAPI   string `json:"openapi"`
	YouTrack  string `json:"youtrack"`
	SwaggerID string `json:"swagger_id"`
	Error     string `json:"error"`
}

func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	var req GenerateReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	document, err := parser.ParseDocument([]byte(req.Text))
	if err != nil {
		DoResponse(w, &GenerateRes{Error: err.Error()})
		return
	}
	if verbose {
		parser.PrettyPrint(document)
	}
	spec, err := gen.GenerateSpec(document)
	if err != nil {
		DoResponse(w, &GenerateRes{Error: err.Error()})
		return
	}

	swaggerID := randomString(10)
	swaggers.Add(Swagger{Key: swaggerID, Spec: spec})

	youtrack, err := markdown.GenMarkdown(document)
	if err != nil {
		DoResponse(w, &GenerateRes{Error: err.Error()})
		return
	}

	DoResponse(w, &GenerateRes{OpenAPI: spec, SwaggerID: swaggerID, YouTrack: youtrack})
}

func DoResponse[T any](w http.ResponseWriter, res *T) {
	data, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(data))
}

type SwaggerBuff struct {
	Data  []Swagger
	Index int
	Len   int
}

func NewSwaggerBuff(l int) *SwaggerBuff {
	return &SwaggerBuff{
		Data:  make([]Swagger, l),
		Index: 0,
		Len:   l,
	}
}

func (b *SwaggerBuff) Add(s Swagger) {
	b.Data[b.Index] = s
	b.Index++
	if b.Index >= b.Len {
		b.Index = 0
	}
}

func (b *SwaggerBuff) Find(key string) (Swagger, bool) {
	for _, s := range b.Data {
		if s.Key == key {
			return s, true
		}
	}
	return Swagger{}, false
}

type Swagger struct {
	Key  string
	Spec string
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

var templateFile = `# Example of api.yml
settings:
  url: /api/v1
  title: Game platform schema
  version: 1.0.0
  security:
    - {}
    - bearer: []

api:
  _common:
    response:
      default: $Error

  user: # Everything about users
    'GET /users/{user_id}': # Return user info
      request:
        params:
          user_id: uuid # User ID (23fb25b8-1780-4bcb-bf28-1a91bb706a54)
      response:
        body:
          data: $User

schemas:
  AnyValue: # Can be anything

  Error:
    code: int64
    message: string
    debug: string?

  Empty:
    data: object

  User:
    id: uuid # (23fb25b8-1780-4bcb-bf28-1a91bb706a54)
    name: string
`
