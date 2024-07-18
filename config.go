package agen

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Kegian/agen/database"
	"github.com/Kegian/agen/sentry"

	"github.com/joho/godotenv"
)

var Config = &BaseConfig{}

type BaseConfig struct {
	Environment string `cfg:"ENVIRONMENT" default:"local"`
	ServerAddr  string `cfg:"SERVER_ADDR" default:":8080"`
	Log         LoggerConfig
	Sentry      sentry.SentryConfig
	Postgres    database.PostgresConfig
	Clickhouse  database.ClickhouseConfig
}

func (c *BaseConfig) Config() *BaseConfig {
	return c
}

type Configurable interface {
	Config() *BaseConfig
}

func LoadConfig(cfg any) error {
	ptr := reflect.ValueOf(cfg)

	if ptr.Kind() != reflect.Pointer || ptr.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("%T should be pointer to a struct to load configs", cfg)
	}

	return loadStructConfig(ptr.Elem())
}

func CurrentConfigs() []string {
	curr := make([]string, 0, len(allConfigsKey))
	for _, key := range allConfigsKey {
		curr = append(curr, fmt.Sprintf("%s=%s", key, allConfigsVal[key].curr))
	}
	return curr
}

func DefaultConfigs() []string {
	def := make([]string, 0, len(allConfigsKey))
	for _, key := range allConfigsKey {
		def = append(def, fmt.Sprintf("%s=%s", key, allConfigsVal[key].def))
	}
	return def
}

func loadStructConfig(elem reflect.Value) error {
	typ := elem.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldElem := elem.Field(i)
		fieldType := typ.Field(i)

		tagCfg := fieldType.Tag.Get("cfg")
		tagDef := fieldType.Tag.Get("default")

		if !fieldType.IsExported() && tagCfg != "" {
			return fmt.Errorf("field '%s' is unexported but have cfg tag", fieldType.Name)
		}

		switch fieldKind := fieldElem.Kind(); {
		case isAllowedType(fieldKind):
			if tagCfg == "" {
				continue
			}

			if err := setField(fieldType.Name, fieldElem, tagCfg, tagDef); err != nil {
				return err
			}

		case fieldKind == reflect.Struct:
			if err := loadStructConfig(fieldElem); err != nil {
				return err
			}

		case fieldKind == reflect.Pointer:
			if fieldElem.IsNil() {
				fieldElem.Set(reflect.New(fieldElem.Type().Elem()))
			}

			switch elemKind := fieldElem.Type().Elem().Kind(); elemKind {
			case reflect.Struct:
				if err := loadStructConfig(fieldElem.Elem()); err != nil {
					return err
				}
			default:
				if tagCfg == "" {
					continue
				}
				if isAllowedType(elemKind) {
					if err := setField(fieldType.Name, fieldElem.Elem(), tagCfg, tagDef); err != nil {
						return err
					}
				}
			}

		default:
			if tagCfg != "" {
				return fmt.Errorf("field '%s' type unsupported but have cfg tag", fieldType.Name)
			}
		}
	}

	return nil
}

var allConfigsKey = []string{}
var allConfigsVal = map[string]cfgValue{}

type cfgValue struct {
	name string
	def  string
	curr string
}

func setField(fieldName string, elem reflect.Value, cfg, def string) error {
	names := strings.Split(cfg, ",")
	value := getenv(names, def)
	if value == "" {
		return nil
	}

	oldValue := value

	if _, ok := allConfigsVal[names[0]]; !ok {
		allConfigsKey = append(allConfigsKey, names[0])
		allConfigsVal[names[0]] = cfgValue{name: names[0], def: def}
	}

	if !elem.IsValid() {
		return fmt.Errorf("field '%s' is not valid", fieldName)
	}

	if !elem.CanSet() {
		return fmt.Errorf("can't set field '%s'", fieldName)
	}

	if reflect.TypeOf(time.Duration(0)) == elem.Type() {
		dur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		value = strconv.Itoa(int(dur))
	}

	switch elem.Kind() {
	case reflect.Bool:
		switch v := strings.ToLower(value); v {
		case "true", "t":
			elem.SetBool(true)
		case "false", "f":
			elem.SetBool(false)
		default:
			return fmt.Errorf("bool field '%s' can't assigne %s", fieldName, value)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			return err
		}
		if elem.OverflowInt(v) {
			return fmt.Errorf("int field '%s' is overflowed with %s", fieldName, value)
		}
		elem.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			return err
		}
		if elem.OverflowUint(v) {
			return fmt.Errorf("uint field '%s' is overflowed with %s", fieldName, value)
		}
		elem.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		if elem.OverflowFloat(v) {
			return fmt.Errorf("float field '%s' is overflowed with %s", fieldName, value)
		}
		elem.SetFloat(v)

	case reflect.String:
		elem.SetString(value)
	}

	allConfigsVal[names[0]] = cfgValue{name: names[0], def: def, curr: oldValue}

	return nil
}

func isAllowedType(kind reflect.Kind) bool {
	switch kind {
	case
		reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

var envs = func() map[string]string {
	e, err := godotenv.Read()
	if err != nil {
		return map[string]string{}
	}
	return e
}()

func getenv(names []string, def string) string {
	for _, name := range names {
		v, ok := os.LookupEnv(name)
		if ok {
			return v
		}
	}

	for _, name := range names {
		if v, ok := envs[name]; ok {
			return v
		}
	}

	return def
}
