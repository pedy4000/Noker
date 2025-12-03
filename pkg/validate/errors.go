package validate

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Format converts validator.ValidationErrors into beautiful, human-readable messages
func Format(err error) string {
	errs, ok := err.(validator.ValidationErrors)
	if !ok || len(errs) == 0 {
		return "Validation failed"
	}

	var b strings.Builder
	b.WriteString("Validation failed: ")

	for i, e := range errs {
		if i > 0 {
			if i == len(errs)-1 {
				b.WriteString(" and ")
			} else {
				b.WriteString(", ")
			}
		}

		field := strings.ToLower(e.Field())

		switch e.Tag() {
		case "required":
			b.WriteString(fmt.Sprintf("'%s' is required", field))
		case "min":
			if e.Type().Kind().String() == "string" {
				b.WriteString(fmt.Sprintf("'%s' must be at least %s characters", field, e.Param()))
			} else {
				b.WriteString(fmt.Sprintf("'%s' must be at least %s", field, e.Param()))
			}
		case "max":
			b.WriteString(fmt.Sprintf("'%s' cannot exceed %s characters", field, e.Param()))
		case "oneof":
			options := strings.ReplaceAll(e.Param(), " ", ", ")
			b.WriteString(fmt.Sprintf("'%s' must be one of: %s", field, options))
		case "email":
			b.WriteString(fmt.Sprintf("'%s' must be a valid email address", field))
		case "uuid":
			b.WriteString(fmt.Sprintf("'%s' must be a valid UUID", field))
		default:
			b.WriteString(fmt.Sprintf("'%s' is invalid", field))
		}
	}

	return b.String()
}
