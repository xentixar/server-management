package validators

import (
	"api/app/constants"

	"github.com/go-playground/validator/v10"
)

func ValidateRole(fl validator.FieldLevel) bool {
	return constants.IsValidRole(fl.Field().String())
}
