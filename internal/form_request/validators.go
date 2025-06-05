package formrequest

import (
	"strings"

	"slices"

	"github.com/go-playground/validator/v10"
)

func ValidateOneOfOrEmpty(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}
	return slices.Contains(strings.Split(fl.Param(), " "), fl.Field().String())
}
