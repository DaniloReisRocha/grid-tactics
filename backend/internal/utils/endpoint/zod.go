package endpoint

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"gopkg.in/guregu/null.v3"
)

type Zod struct {
	Schemas  string
	Endpoint string
}

func MakeZod[B any, P any, Q any, R any](method, path string) (zod *Zod, err error) {
	prefix := buildPrefix(method, path)

	var sb strings.Builder
	var generatedSchemas []string

	sb.WriteString(fmt.Sprintf(`{
  method: "%s",
  path: "%s",
  alias: "%s",
`, strings.ToLower(method), path, buildAlias(method, path)))

	var paramSchemas []string

	var params P
	if !isEmpty(params) {
		schemaName := prefix + "_params_schema"
		paramSchemas = append(paramSchemas, fmt.Sprintf(`{ name: "params", type: "Path", schema: %s }`, schemaName))
		generatedSchemas = append(generatedSchemas, ToZodSchema(params, schemaName))
	}

	var query Q
	if !isEmpty(query) {
		schemaName := prefix + "_query_schema"
		paramSchemas = append(paramSchemas, fmt.Sprintf(`{ name: "query", type: "Query", schema: %s }`, schemaName))
		generatedSchemas = append(generatedSchemas, ToZodSchema(query, schemaName))
	}

	var body B
	if !isEmpty(body) {
		schemaName := prefix + "_body_schema"
		paramSchemas = append(paramSchemas, fmt.Sprintf(`{ name: "body", type: "Body", schema: %s }`, schemaName))
		generatedSchemas = append(generatedSchemas, ToZodSchema(body, schemaName))
		sb.WriteString("  requestFormat: \"json\",\n")
	}

	if len(paramSchemas) > 0 {
		sb.WriteString("  parameters: [\n")
		for _, p := range paramSchemas {
			sb.WriteString("    " + p + ",\n")
		}
		sb.WriteString("  ],\n")
	}

	var response R
	if !isEmpty(response) {
		schemaName := prefix + "_response_schema"
		generatedSchemas = append(generatedSchemas, ToZodSchema(response, schemaName))
		sb.WriteString(fmt.Sprintf("  response: %s,\n", schemaName))
	} else {
		sb.WriteString("  response: z.any(),\n")
	}

	sb.WriteString("},\n")

	return &Zod{
		Schemas:  strings.Join(generatedSchemas, "\n"),
		Endpoint: sb.String(),
	}, nil
}

func isEmpty(v interface{}) bool {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Struct && t.NumField() == 0 {
		return true
	}
	return false
}

func ToZodSchema(input interface{}, typeName string) string {
	t := reflect.TypeOf(input)
	if t.Kind() != reflect.Struct {
		return fmt.Sprintf("// %s: Tipo não suportado (não é struct)", typeName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export const %s = z.object({\n", typeName))

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" || fieldName == "-" {
			fieldName = strings.ToLower(field.Name[:1]) + field.Name[1:]
		}
		zodType := mapGoTypeToZod(field.Type, make(map[reflect.Type]bool))
		sb.WriteString(fmt.Sprintf("  %s: %s,\n", fieldName, zodType))
	}

	sb.WriteString("});")
	return sb.String()
}

func mapGoTypeToZod(t reflect.Type, visited map[reflect.Type]bool) string {
	textMarshalerType := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	jsonMarshalerType := reflect.TypeOf((*json.Marshaler)(nil)).Elem()

	if t.Kind() == reflect.Ptr {
		return mapGoTypeToZod(t.Elem(), visited)
	}

	if visited[t] {
		return "z.lazy(() => z.any())"
	}
	visited[t] = true
	defer delete(visited, t)

	if t.Implements(textMarshalerType) || reflect.PointerTo(t).Implements(textMarshalerType) ||
		t.Implements(jsonMarshalerType) || reflect.PointerTo(t).Implements(jsonMarshalerType) {
		return "z.string()"
	}

	switch t.Kind() {
	case reflect.String:
		return "z.string()"
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
		reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
		reflect.Float64, reflect.Float32:
		return "z.number()"
	case reflect.Bool:
		return "z.boolean()"
	case reflect.Slice, reflect.Array:
		return fmt.Sprintf("z.array(%s)", mapGoTypeToZod(t.Elem(), visited))
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			return fmt.Sprintf("z.record(%s)", mapGoTypeToZod(t.Elem(), visited))
		}
		return "z.any()"
	case reflect.Struct:
		fields := ""
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			if !field.IsExported() {
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}
			fieldName := field.Name
			if jsonTag != "" {
				if tagName := strings.Split(jsonTag, ",")[0]; tagName != "" {
					fieldName = tagName
				}
			}

			fieldZod := mapGoTypeToZod(field.Type, visited)

			if isNullableType(field.Type) {
				fieldZod += ".nullable()"
			}

			if strings.Contains(jsonTag, "omitempty") {
				fieldZod += ".optional()"
			}

			fields += fmt.Sprintf("%s: %s,\n", fieldName, fieldZod)
		}
		return fmt.Sprintf("z.object({\n%s})", fields)
	default:
		return "z.any()"
	}
}

func isNullableType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		return true
	}
	nullables := []reflect.Type{
		reflect.TypeOf(null.Time{}),
	}

	for _, nt := range nullables {
		if t == nt {
			return true
		}
	}

	return false
}

func buildPrefix(method, path string) string {
	var sb strings.Builder

	sb.WriteString(strings.ToLower(method))
	sb.WriteString("_")

	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, ":") {
			sb.WriteString("By" + part[1:])
		} else {
			sb.WriteString(part)
		}
		sb.WriteString("_")
	}

	prefix := sb.String()
	prefix = strings.TrimSuffix(prefix, "_")
	prefix = strings.ReplaceAll(prefix, "-", "_")

	return prefix
}

func buildAlias(method, path string) string {
	path = strings.TrimPrefix(path, "/api/")
	parts := strings.Split(path, "/")

	var aliasParts []string
	for _, part := range parts {
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, ":") {
			aliasParts = append(aliasParts, "by_"+part[1:])
		} else {
			aliasParts = append(aliasParts, part)
		}
	}
	aliasParts = append(aliasParts, strings.ToLower(method))

	return toCamelCase(strings.Join(aliasParts, "_"))
}

func toCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	for i := range words {
		words[i] = strings.ToLower(words[i])
		if i > 0 && len(words[i]) > 0 {
			words[i] = strings.ToUpper(string(words[i][0])) + words[i][1:]
		}
	}

	return strings.Join(words, "")
}
