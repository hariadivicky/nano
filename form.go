package nano

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// BindSimpleForm is functions to bind request body (with content type form-urlencoded or url query) to targetStruct.
// targetStruct must be pointer to user defined struct.
func BindSimpleForm(r *http.Request, targetStruct interface{}) *BindingError {
	// only accept pointer
	if reflect.TypeOf(targetStruct).Kind() != reflect.Ptr {
		return &BindingError{
			Message:        "expected pointer to target struct, got non-pointer",
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	err := r.ParseForm()
	if err != nil {
		return &BindingError{
			Message:        fmt.Sprintf("could not parsing form body: %v", err),
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	err = bindForm(r.Form, targetStruct)
	if err != nil {
		return &BindingError{
			HTTPStatusCode: http.StatusInternalServerError,
			Message:        fmt.Sprintf("binding error: %v", err),
		}
	}

	return ValidateStruct(targetStruct)
}

// BindMultipartForm is functions to bind request body (with contet type multipart/form-data) to targetStruct.
// targetStruct must be pointer to user defined struct.
func BindMultipartForm(r *http.Request, targetStruct interface{}) *BindingError {
	// only accept pointer
	if reflect.TypeOf(targetStruct).Kind() != reflect.Ptr {
		return &BindingError{
			Message:        "expected pointer to target struct, got non-pointer",
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	err := r.ParseMultipartForm(16 << 10)
	if err != nil {
		return &BindingError{
			Message:        fmt.Sprintf("could not parsing form body: %v", err),
			HTTPStatusCode: http.StatusBadRequest,
		}
	}

	err = bindForm(r.MultipartForm.Value, targetStruct)
	if err != nil {
		return &BindingError{
			HTTPStatusCode: http.StatusInternalServerError,
			Message:        fmt.Sprintf("binding error: %v", err),
		}
	}

	return ValidateStruct(targetStruct)
}

// bindForm will map each field in request body into targetStruct.
func bindForm(form map[string][]string, targetStruct interface{}) error {
	targetPtr := reflect.ValueOf(targetStruct).Elem()
	targetType := targetPtr.Type()

	// only accept struct as target binding
	if targetPtr.Kind() != reflect.Struct {
		return fmt.Errorf("expected target binding to be struct")
	}

	for i := 0; i < targetPtr.NumField(); i++ {
		fieldValue := targetPtr.Field(i)
		// this is used to get field tag.
		fieldType := targetType.Field(i)

		// continue iteration when field is not settable.
		if !fieldValue.CanSet() {
			continue
		}

		// check if current field nested struct.
		// this is possible when current request body is json type.
		if fieldValue.Kind() == reflect.Struct {
			// bind recursively.
			err := bindForm(form, fieldValue.Addr().Interface())
			if err != nil {
				return err
			}
		} else {
			// web use tag "form" as field name in request body.
			// so make sure you have matching name at field name in request body and field tag in your target struct
			formFieldName := fieldType.Tag.Get("form")
			// continue iteration when field doesnt have form tag.
			if formFieldName == "" {
				continue
			}

			formValue, exists := form[formFieldName]
			// could not find value in request body, let it empty
			if !exists {
				continue
			}

			formValueCount := len(formValue)
			// it's possible if current field value is an array.
			if fieldValue.Kind() == reflect.Slice && formValueCount > 0 {
				sliceKind := fieldValue.Type().Elem().Kind()
				slice := reflect.MakeSlice(fieldValue.Type(), formValueCount, formValueCount)
				for i := 0; i < formValueCount; i++ {
					if err := setFieldValue(sliceKind, formValue[i], slice.Index(i)); err != nil {
						return err
					}
				}
				fieldValue.Field(i).Set(slice)
			} else {
				// it's a single value. just do direct set.
				if err := setFieldValue(fieldValue.Kind(), formValue[0], fieldValue); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// setFieldValue is functions to set field with typed value.
// we will find the best type & size for your field value.
// if empty string provided to value parameter, we will use zero type value as default field value.
func setFieldValue(kind reflect.Kind, value string, fieldValue reflect.Value) error {
	switch kind {
	case reflect.Int:
		setIntField(value, 0, fieldValue)
	case reflect.Int8:
		setIntField(value, 8, fieldValue)
	case reflect.Int16:
		setIntField(value, 16, fieldValue)
	case reflect.Int32:
		setIntField(value, 32, fieldValue)
	case reflect.Int64:
		setIntField(value, 64, fieldValue)
	case reflect.Uint:
		setUintField(value, 0, fieldValue)
	case reflect.Uint8:
		setUintField(value, 8, fieldValue)
	case reflect.Uint16:
		setUintField(value, 16, fieldValue)
	case reflect.Uint32:
		setUintField(value, 32, fieldValue)
	case reflect.Uint64:
		setUintField(value, 64, fieldValue)
	case reflect.Bool:
		setBoolField(value, fieldValue)
	case reflect.Float32:
		setFloatField(value, 32, fieldValue)
	case reflect.Float64:
		setFloatField(value, 64, fieldValue)
	case reflect.String:
		// no conversion needed. because value already a string.
		fieldValue.SetString(value)
	default:
		// whoopss..
		return fmt.Errorf("unknown type")
	}
	return nil
}

// setIntField is functions to convert input string (value) into integer.
func setIntField(value string, size int, field reflect.Value) {
	convertedValue, err := strconv.ParseInt(value, 10, size)
	// set default empty value when conversion.
	if err != nil {
		convertedValue = 0
	}
	field.SetInt(convertedValue)
}

// setUintField is functions to convert input string (value) into unsigned integer.
func setUintField(value string, size int, field reflect.Value) {
	convertedValue, err := strconv.ParseUint(value, 10, size)
	// set default empty value when conversion.
	if err != nil {
		convertedValue = 0
	}
	field.SetUint(convertedValue)
}

// setBoolField is functions to convert input string (value) into boolean.
func setBoolField(value string, field reflect.Value) {
	convertedValue, err := strconv.ParseBool(value)
	// set default empty value when conversion.
	if err != nil {
		convertedValue = false
	}
	field.SetBool(convertedValue)
}

// setFloatField is functions to convert input string (value) into floating.
func setFloatField(value string, size int, field reflect.Value) {
	convertedValue, err := strconv.ParseFloat(value, size)
	// set default empty value when conversion.
	if err != nil {
		convertedValue = 0.0
	}
	field.SetFloat(convertedValue)
}
