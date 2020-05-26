package nano

import (
	"log"
	"net/http"
	"testing"
)

func TestBindJSON(t *testing.T) {
	type Person struct {
		Name   string
		Gender string
	}

	var person Person

	t.Run("bind non pointer struct", func(st *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			log.Fatalf("could not make http request: %v", err)
		}
		errBinding := BindJSON(req, person)

		if errBinding == nil {
			st.Errorf("expected error to be returned; got %v", errBinding)
		}

		if errBinding != &ErrBindNonPointer {
			st.Errorf("expect error to be ErrBindNonPointer; got %v", errBinding)
		}
	})
}
