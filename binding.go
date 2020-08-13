package nano

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// BindingError is an error wrapper.
// HTTPStatusCode will set to 422 when there is error on validation,
// 400 when client sent unsupported/without Content-Type header, and
// 500 when targetStruct is not pointer or type conversion is fail.
type BindingError struct {
	HTTPStatusCode int
	Message        string
	Fields         []string
}

var (
	// ErrBindNonPointer must be returned when non-pointer struct passed as targetStruct parameter.
	ErrBindNonPointer = &BindingError{
		Message:        "expected pointer to target struct, got non-pointer",
		HTTPStatusCode: http.StatusInternalServerError,
	}

	// ErrBindContentType returned when client content type besides json, urlencoded, & multipart form.
	ErrBindContentType = &BindingError{
		HTTPStatusCode: http.StatusBadRequest,
		Message:        "unknown content type of request body",
	}
)

// Bind request body into defined user struct.
// This function help you to automatic binding based on request Content-Type & request method.
// If you want to chooose binding method manually, you could use :
// BindSimpleForm to bind urlencoded form & url query,
// BindMultipartForm to bind multipart/form data,
// and BindJSON to bind application/json request body.
func (c *Context) Bind(targetStruct interface{}) *BindingError {
	contentType := c.GetRequestHeader(HeaderContentType)

	// if client request using POST, PUT, & PATCH we will try to bind request using simple form (urlencoded & url query),
	// multipart form, and JSON. if you need both binding e.g. to bind multipart form & url query,
	// this method doesn't works. you should call BindSimpleForm & BindMultipartForm manually from your handler.
	if c.Method == http.MethodPost || c.Method == http.MethodPut || c.Method == http.MethodPatch || contentType != "" {
		if strings.Contains(contentType, MimeFormURLEncoded) {
			return c.BindSimpleForm(targetStruct)
		}

		if strings.Contains(contentType, MimeMultipartForm) {
			return c.BindMultipartForm(targetStruct)
		}

		if c.IsJSON() {
			return c.BindJSON(targetStruct)
		}

		return ErrBindContentType
	}

	// when client request using GET method, we will serve binding using simple form.
	// it's can binding url-encoded form & url query data.
	return c.BindSimpleForm(targetStruct)
}

// BindJSON is functions to bind request body (with contet type application/json) to targetStruct.
// targetStruct must be pointer to user defined struct.
func (c *Context) BindJSON(targetStruct interface{}) *BindingError {
	// only accept pointer
	if reflect.TypeOf(targetStruct).Kind() != reflect.Ptr {
		return ErrBindNonPointer
	}

	if c.Request.Body != nil {
		defer c.Request.Body.Close()
		err := json.NewDecoder(c.Request.Body).Decode(targetStruct)
		if err != nil && err != io.EOF {
			return &BindingError{
				Message:        err.Error(),
				HTTPStatusCode: http.StatusBadRequest,
			}
		}
	}

	return validate(c, targetStruct)
}

// BindSimpleForm is functions to bind request body (with content type form-urlencoded or url query) to targetStruct.
// targetStruct must be pointer to user defined struct.
func (c *Context) BindSimpleForm(targetStruct interface{}) *BindingError {
	// only accept pointer
	if reflect.TypeOf(targetStruct).Kind() != reflect.Ptr {
		return &BindingError{
			Message:        "expected pointer to target struct, got non-pointer",
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	err := c.Request.ParseForm()
	if err != nil {
		return &BindingError{
			Message:        fmt.Sprintf("could not parsing form body: %v", err),
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	err = bindForm(c.Request.Form, targetStruct)
	if err != nil {
		return &BindingError{
			HTTPStatusCode: http.StatusInternalServerError,
			Message:        fmt.Sprintf("binding error: %v", err),
		}
	}

	return validate(c, targetStruct)
}

// BindMultipartForm is functions to bind request body (with contet type multipart/form-data) to targetStruct.
// targetStruct must be pointer to user defined struct.
func (c *Context) BindMultipartForm(targetStruct interface{}) *BindingError {
	// only accept pointer
	if reflect.TypeOf(targetStruct).Kind() != reflect.Ptr {
		return &BindingError{
			Message:        "expected pointer to target struct, got non-pointer",
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	err := c.Request.ParseMultipartForm(16 << 10)
	if err != nil {
		return &BindingError{
			Message:        fmt.Sprintf("could not parsing form body: %v", err),
			HTTPStatusCode: http.StatusBadRequest,
		}
	}

	err = bindForm(c.Request.MultipartForm.Value, targetStruct)
	if err != nil {
		return &BindingError{
			HTTPStatusCode: http.StatusInternalServerError,
			Message:        fmt.Sprintf("binding error: %v", err),
		}
	}

	return validate(c, targetStruct)
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
