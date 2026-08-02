package inference

import (
	"encoding/json"

	"github.com/EmptyZeroRain/json2struct/schema"
)

// Infer creates a schema from a decoded JSON value.
func Infer(value interface{}) *schema.Field {
	return infer("", value)
}

func infer(name string, value interface{}) *schema.Field {
	f := schema.NewField(name, schema.TypeUnknown)
	switch v := value.(type) {
	case nil:
		f.Type, f.Nullable = schema.TypeNull, true
	case string:
		f.Type = schema.TypeString
	case bool:
		f.Type = schema.TypeBoolean
	case json.Number:
		f.Type = schema.TypeString
	case float64:
		f.Type = schema.TypeString
	case map[string]interface{}:
		f.Type, f.Children = schema.TypeObject, map[string]*schema.Field{}
		for k, child := range v {
			f.Children[k] = infer(k, child)
		}
	case []interface{}:
		f.Type, f.Array = schema.TypeArray, true
		if len(v) > 0 {
			f.Element = infer("", v[0])
			for _, item := range v[1:] {
				f.Element = schema.Merge(f.Element, infer("", item))
			}
		}
	default:
		// JSON normally only produces the cases above. Keep the inference
		// conservative for values supplied directly to Infer: an unrecognised
		// value is represented as a string instead of leaking interface{} into
		// generated models.
		f.Type = schema.TypeString
	}
	return f
}
