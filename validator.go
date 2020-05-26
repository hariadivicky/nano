package nano

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

var (
	// ErrBindNonPointer must be returned when non-pointer struct passed as targetStruct parameter.
	ErrBindNonPointer = BindingError{
		Message:        "expected pointer to target struct, got non-pointer",
		HTTPStatusCode: http.StatusInternalServerError,
	}
)

// ValidateStruct will call default struct validator and collect error information from each struct field.
func ValidateStruct(targetStruct interface{}) *BindingError {
	// only accept pointer
	if reflect.TypeOf(targetStruct).Kind() != reflect.Ptr {
		return &ErrBindNonPointer
	}

	errorBag := make([]string, 0)

	// collect error from each field.
	errorFields := validate(targetStruct, errorBag)
	if len(errorFields) > 0 {
		// if only two error fields, just join with & for better message.
		if len(errorFields) == 2 {
			return &BindingError{
				HTTPStatusCode: http.StatusUnprocessableEntity,
				Message:        fmt.Sprintf("%s fields are required", strings.Join(errorFields, " & ")),
			}
		}

		return &BindingError{
			HTTPStatusCode: http.StatusUnprocessableEntity,
			Message:        fmt.Sprintf("%s fields are required", strings.Join(errorFields, ", ")),
		}
	}

	return nil
}

// validate is default struct validator. this function will called when you do request binding to some struct.
// Current validation rule is only to validate "required" field. To apply field into validation, just add "rules" at field tag.
// if you apply "required" rule, that is mean you are not allowed to use zero type value in you request body field
// because it will give you validation error.
// so if you need 0 value for int field or false value for boolean field, pelase consider to not use "required" rules.
func validate(targetStruct interface{}, errFields []string) []string {
	targetValue := reflect.ValueOf(targetStruct)
	if reflect.TypeOf(targetStruct).Kind() == reflect.Ptr {
		targetValue = targetValue.Elem()
	}

	targetType := targetValue.Type()

	for i := 0; i < targetType.NumField(); i++ {
		fieldType := targetType.Field(i)

		// skip ignored and unexported fields in the struct
		if fieldType.Tag.Get("form") == "-" || !targetValue.Field(i).CanInterface() {
			continue
		}

		// get real values of field and zero type value.
		fieldValue := targetValue.Field(i).Interface()
		zeroValue := reflect.Zero(fieldType.Type).Interface()

		// validate nested struct inside field.
		if fieldType.Type.Kind() == reflect.Struct && !reflect.DeepEqual(zeroValue, fieldValue) {
			errFields = validate(fieldValue, errFields)
		}

		// only validate when tag rules is set to required.
		if strings.Contains(fieldType.Tag.Get("rules"), "required") {
			if reflect.DeepEqual(zeroValue, fieldValue) {
				// use field name as default when json & form tag both are not provided.
				name := strings.ToLower(fieldType.Name)

				// try to get input name from json tag.
				// if field has both of json & form tag, appended field name in errFields will taken from form tag.
				if jsonName := fieldType.Tag.Get("json"); jsonName != "" {
					name = jsonName
				}

				// try to get input name from form tag.
				if formName := fieldType.Tag.Get("form"); formName != "" {
					name = formName
				}

				errFields = append(errFields, name)
			}
		}
	}

	return errFields
}
