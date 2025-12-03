package validate

import (
	"github.com/go-playground/validator/v10"
)

// Single global validator instance (thread-safe)
var validatorInstance = validator.New(
	validator.WithRequiredStructEnabled(), // treats omitted fields as required if struct tag has `required`
)

// Do validates any struct using struct tags
func Do(s any) error {
	return validatorInstance.Struct(s)
}
