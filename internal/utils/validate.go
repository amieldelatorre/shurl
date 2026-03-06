package utils

import (
	"log/slog"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	err      error
	doOnce   sync.Once
)

func ValidateLogLevel(field validator.FieldLevel) bool {
	var l slog.Level
	err := l.UnmarshalText([]byte(field.Field().String()))
	return err == nil
}

func GetValidator() (*validator.Validate, error) {
	doOnce.Do(func() {
		validate = validator.New()
		err = validate.RegisterValidation("loglevelvalidator", ValidateLogLevel)
	})
	return validate, err
}
