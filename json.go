package nano

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

// BindJSON is functions to bind request body (with contet type application/json) to targetStruct.
// targetStruct must be pointer to user defined struct.
func BindJSON(r *http.Request, targetStruct interface{}) *BindingError {
	// only accept pointer
	if reflect.TypeOf(targetStruct).Kind() != reflect.Ptr {
		return &ErrBindNonPointer
	}

	if r.Body != nil {
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(targetStruct)
		if err != nil && err != io.EOF {
			return &BindingError{
				Message:        err.Error(),
				HTTPStatusCode: http.StatusBadRequest,
			}
		}
	}

	return ValidateStruct(targetStruct)
}
