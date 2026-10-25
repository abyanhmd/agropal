package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator so struct-tag validation failures
// report each field by its `json` tag (the name the API client actually
// sent) instead of the Go struct field name.
type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	return &Validator{validate: v}
}

// ValidateStruct runs struct-tag validation over s and returns a field ->
// messages map describing every failure, or nil if s is valid.
func (v *Validator) ValidateStruct(s any) ValidationErrors {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrs validator.ValidationErrors
	if !errors.As(err, &fieldErrs) {
		// Not a validation failure (e.g. s wasn't a struct) - nothing
		// field-level to report.
		return nil
	}

	out := make(ValidationErrors, len(fieldErrs))
	for _, fe := range fieldErrs {
		out[fe.Field()] = append(out[fe.Field()], validationMessage(fe))
	}
	return out
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}
