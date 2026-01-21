package dto

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	Validate = validator.New()
	trans    ut.Translator
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type Response struct {
	Message string `json:"message"`
}

func InitValidator() error {
	uni := ut.New(en.New(), en.New())
	trans, _ = uni.GetTranslator("en")

	err := enTranslations.RegisterDefaultTranslations(Validate, trans)
	if err != nil {
		return err
	}

	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return nil
}

func ValidateSingleError(req interface{}) error {
	if err := Validate.Struct(req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return errors.New(ve[0].Translate(trans))
		}
		return err
	}
	return nil
}
