package nano

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupContext() *Context {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	return newContext(w, r)

}

func TestValidator(t *testing.T) {
	type Person struct {
		Name         string `form:"name" json:"name" validate:"required"`
		Gender       string `form:"gender" json:"gender" validate:"required"`
		Email        string `form:"email" json:"email" validate:"required"`
		Phone        string `form:"phone" json:"phone"`
		privateField string
		IgnoredField string `form:"-"`
	}

	person := Person{
		Name:   "foo",
		Gender: "male",
		Email:  "hariadivicky@gmail.com",
		Phone:  "",
	}

	ctx := setupContext()

	t.Run("pass non-pointer struct", func(st *testing.T) {
		err := validate(ctx, person)
		if err == nil {
			t.Fatalf("expected error to be returned")
		}

		if errBind, ok := err.(ErrBinding); ok {
			if errBind.Status != ErrBindNonPointer.Status {
				st.Errorf("expected HTTPStatusCode error to be %d; got %d", ErrBindNonPointer.Status, errBind.Status)
			}

			if errBind.Text != ErrBindNonPointer.Text {
				st.Errorf("expected error message to be %s; got %s", ErrBindNonPointer.Text, errBind.Text)
			}
		}
	})

	t.Run("validation should be passed", func(st *testing.T) {
		errBind := validate(ctx, &person)

		if errBind != nil {
			t.Errorf("expected error binding to be nil; got %v", errBind)
		}
	})

	t.Run("empty value on required fields", func(st *testing.T) {
		person.Name = ""
		person.Gender = ""
		person.Email = ""

		err := validate(ctx, &person)
		if err == nil {
			st.Fatalf("expected error to be returned")
		}

		if bindErr, ok := err.(ErrBinding); ok {
			if bindErr.Status != http.StatusUnprocessableEntity {
				st.Errorf("expected HTTPStatusCode error to be %d; got %d", ErrBindNonPointer.Status, http.StatusUnprocessableEntity)
			}

			if bindErr.Text != "validation error" {
				st.Errorf("expected error message to be %s; got %s", ErrBindNonPointer.Text, bindErr.Text)
			}

			if errFieldsCount := len(bindErr.Fields); errFieldsCount != 3 {
				st.Fatalf("expected num of error fields to be 3; got %d", errFieldsCount)
			}

			errFields := []string{
				"name is a required field",
				"gender is a required field",
				"email is a required field",
			}

			for i, errMsg := range bindErr.Fields {
				if errMsg != errFields[i] {
					st.Errorf("expected error %d to be %s; got %s", i, errFields[i], errMsg)
				}
			}

			return
		}

		st.Fatalf("expected error type to be ErrBinding, got %T", err)
	})

}

func TestNestedStructValidation(t *testing.T) {
	type Person struct {
		Name    string `form:"name" json:"name" validate:"required"`
		Gender  string `form:"gender" json:"gender" validate:"required"`
		Address struct {
			CityID     int `form:"city_id" json:"city_id" validate:"required"`
			PostalCode int `form:"postal_code" json:"postal_code"`
		}
	}

	person := Person{
		Name:   "foo",
		Gender: "",
		Address: struct {
			CityID     int `form:"city_id" json:"city_id" validate:"required"`
			PostalCode int `form:"postal_code" json:"postal_code"`
		}{
			CityID:     0,
			PostalCode: 204,
		},
	}

	ctx := setupContext()

	err := validate(ctx, &person)
	if err == nil {
		t.Fatalf("expected error to be returned")
	}

	if errBind, ok := err.(ErrBinding); ok {
		if errBind.Status != http.StatusUnprocessableEntity {
			t.Errorf("expected error HTTPStatusCode to be %d; got %d", http.StatusUnprocessableEntity, errBind.Status)
		}

		if errBind.Text != "validation error" {
			t.Errorf("expected error message to be validation error; got %s", errBind.Text)
		}

		errFields := []string{
			"gender is a required field",
			"city_id is a required field",
		}

		for i, errMsg := range errBind.Fields {
			if errMsg != errFields[i] {
				t.Errorf("expected error %d to be %s; got %s", i, errFields[i], errMsg)
			}
		}

		return
	}

	t.Fatalf("expected ErrBinding, got %T", err)
}
