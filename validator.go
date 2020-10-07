package nano

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// newTranslator returns validator translation. default using "en"
func newTranslator() ut.Translator {
	// NOTE: ommitting allot of error checking for brevity
	en := en.New()
	uni := ut.New(en, en)

	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	trans, _ := uni.GetTranslator("en")
	return trans
}

func newValidator(trans ut.Translator) *validator.Validate {
	v10 := validator.New()
	v10.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	en_translations.RegisterDefaultTranslations(v10, trans)
	return v10
}

// validate is default struct validator. this function will called when you do request binding to some struct.
// Current validation rule is only to validate "required" field. To apply field into validation, just add "rules" at field tag.
// if you apply "required" rule, that is mean you are not allowed to use zero type value in you request body field
// because it will give you validation error.
// so if you need 0 value for int field or false value for boolean field, pelase consider to not use "required" rules.
func validate(c *Context, targetStruct interface{}) error {
	// only accept pointer
	if reflect.TypeOf(targetStruct).Kind() != reflect.Ptr {
		return &ErrBinding{
			Text:   "expected pointer to target struct, got non-pointer",
			Status: http.StatusInternalServerError,
		}
	}

	err := c.validator.Struct(targetStruct)

	if err != nil {
		var errFields []string
		for _, err := range err.(validator.ValidationErrors) {
			errFields = append(errFields, err.Translate(c.translator))
		}

		return ErrBinding{
			Status: http.StatusUnprocessableEntity,
			Text:   "validation error",
			Fields: errFields,
		}
	}

	return nil
}
