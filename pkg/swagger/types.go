package swagger

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/seniorGolang/i2s/pkg/node"
)

func (b *Builder) buildTypes(node node.Node, swagger *Swagger) {

	if swagger.Components.Schemas == nil {
		swagger.Components.Schemas = make(schemas)
	}

	for _, object := range node.Types {
		swagger.Components.Schemas[object.Type] = b.makeType(object, swagger)
	}

	for _, service := range node.Services {

		for _, method := range service.Methods {

			for _, object := range method.Arguments {

				if len(object.Fields) > 0 {
					swagger.Components.Schemas[object.Type] = b.makeType(object, swagger)
				}
			}

			for _, object := range method.Results {

				if len(object.Fields) > 0 {
					swagger.Components.Schemas[object.Type] = b.makeType(object, swagger)
				}
			}
		}
	}
	return
}

func (b *Builder) makeType(object *node.Object, swagger *Swagger) (com schema) {

	com.Nullable = object.IsNullable
	com.Description = object.Tags.Value("desc")
	com.Example = object.Value(object.Tags.Value("example"))

	typeName, format := castType(object)

	if object.Type == "Interface" {

		com.Type = "object"
		if err := json.Unmarshal([]byte(object.Tags.Value("example", "{}")), &com.Example); err != nil {
			log.Error(err)
		}
		return
	}

	if len(object.Fields) != 0 {

		com.Type = "object"

		com.Properties = make(map[string]schema)

		for _, field := range object.Fields {

			if jsonTags, found := field.TypeTags["json"]; found {
				field.Name = jsonTags[0]
			}

			if !field.IsPrivate && field.Name != "-" {
				com.Properties[field.Name] = b.makeType(field, swagger)
			}
		}

		if object.Alias != "-" {
			swagger.Components.Schemas[object.Type] = com
		}
	}

	if object.IsArray {

		com.Type = "array"
		com.Properties = nil

		if format == "binary" {

			com.Type = typeName
			com.Format = format

		} else if object.IsBuildIn {
			com.Items = &schema{Type: typeName, Format: format}
		} else {
			com.Items = &schema{Type: "", Format: format, Ref: fmt.Sprintf("#/components/schemas/%s", typeName)}
		}
		return
	}

	if object.IsMap {

		com.Type = "object"

		if len(object.SubTypes["value"].Fields) != 0 {

			b.makeType(object.SubTypes["value"], swagger)

			com.AdditionalProperties = map[string]string{
				"$ref": fmt.Sprintf("#/components/schemas/%s", object.SubTypes["value"].Type),
			}

		} else if object.SubTypes["value"].Alias != "" {

			com.AdditionalProperties = map[string]string{
				"$ref": fmt.Sprintf("#/components/schemas/%s", object.SubTypes["value"].Alias),
			}

		} else {

			valueType, _ := castType(object.SubTypes["value"])
			com.AdditionalProperties = map[string]string{
				"type": valueType,
			}
		}

		if err := json.Unmarshal([]byte(object.Tags.Value("example", "{}")), &com.Example); err != nil {
			log.Error(err)
		}
		return
	}

	if len(object.Fields) == 0 {

		if object.IsArray {
			com.Items = &schema{Type: typeName, Format: format}
			com.Type = "array"
			return
		}

		if object.Alias != "" && object.Alias != "-" {
			com.Ref = fmt.Sprintf("#/components/schemas/%s", object.Alias)
			return
		}

		com.Type = typeName
		com.Format = format
		return
	}
	return
}

func castType(object *node.Object) (typeName, format string) {

	typeName = object.Type

	switch typeName {

	case "bool":
		typeName = "boolean"

	case "Interface":
		typeName = "object"

	case "time.Time":
		format = "date-time"
		typeName = "string"

	case "byte":
		format = "byte"
		typeName = "string"

		if object.IsArray {
			format = "binary"
		}

	case "uuid.UUID":
		format = "uuid"
		typeName = "string"

	case "float32", "float64":
		format = "float"
		typeName = "number"

	case "int", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		typeName = "number"

	default:
		if object.IsArray {
			typeName = strings.TrimPrefix(object.Type, "[]")
		}
	}

	format = object.Tags.Value("format", format)
	typeName = object.Tags.Value("type", typeName)
	return
}
