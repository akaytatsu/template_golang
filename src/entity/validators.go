package entity

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type IError struct {
	Field string
	Tag   string
	Value string
}

var validate *validator.Validate = validator.New()

// ValidateRawPassword valida uma senha em texto puro contra as mesmas regras
// declaradas na tag `validate` de EntityUser.Password.
//
// A validação é feita antes do bcrypt: checar o hash (sempre 60 caracteres)
// não diria nada sobre a senha que o usuário escolheu.
func ValidateRawPassword(raw string) error {
	return validate.Var(raw, "required,min=4,max=120")
}

// GetStructError traduz um erro de validação numa lista de campos com problema.
// Retorna nil para erros que não vêm do validator.
func GetStructError(err error) []IError {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}

	fieldErrors := make([]IError, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		fieldErrors = append(fieldErrors, IError{
			Field: fieldErr.Field(),
			Tag:   fieldErr.Tag(),
			// %v e não uma type assertion: o campo pode não ser string
			// (IsAdmin é bool, IDs é []uint).
			Value: fmt.Sprintf("%v", fieldErr.Value()),
		})
	}

	return fieldErrors
}
