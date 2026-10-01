package validators

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationErrors(err error, obj any) map[string]string {
	out := make(map[string]string)

	if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
		t := reflect.TypeOf(obj)

		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		for _, fe := range ve {
			fieldName := fe.Field()
			if field, exists := t.FieldByName(fe.StructField()); exists {
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" && jsonTag != "-" {
					fieldName = strings.Split(jsonTag, ",")[0]
				}
			}

			out[fieldName] = msgForTag(fe.Tag(), fe.Param())
		}
		return out
	}

	out["error"] = "Invalid request payload format"
	return out
}

func msgForTag(tag string, param string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address format"
	case "role":
		return "This role is invalid"
	case "min":
		return fmt.Sprintf("Must be at least %s characters long", param)
	default:
		return fmt.Sprintf("Validation failed on rule: %s", tag)
	}
}
