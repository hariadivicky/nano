package nano

import (
	"net/http"
	"testing"
)

func TestValidator(t *testing.T) {
	type Person struct {
		Name         string `form:"name" json:"name" rules:"required"`
		Gender       string `form:"gender" json:"gender" rules:"required"`
		Email        string `form:"email" json:"email" rules:"required"`
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

	t.Run("pass non-pointer struct", func(st *testing.T) {
		bindErr := ValidateStruct(person)

		if bindErr.HTTPStatusCode != ErrBindNonPointer.HTTPStatusCode {
			st.Errorf("expected HTTPStatusCode error to be %d; got %d", ErrBindNonPointer.HTTPStatusCode, bindErr.HTTPStatusCode)
		}

		if bindErr.Message != ErrBindNonPointer.Message {
			st.Errorf("expected error message to be %s; got %s", ErrBindNonPointer.Message, bindErr.Message)
		}
	})

	t.Run("validation should be passed", func(st *testing.T) {
		errBind := ValidateStruct(&person)

		if errBind != nil {
			t.Errorf("expected error binding to be nil; got %v", errBind)
		}
	})

	t.Run("empty value on required fields", func(st *testing.T) {
		person.Name = ""

		bindErr := ValidateStruct(&person)

		if bindErr.HTTPStatusCode != http.StatusUnprocessableEntity {
			st.Errorf("expected HTTPStatusCode error to be %d; got %d", ErrBindNonPointer.HTTPStatusCode, http.StatusUnprocessableEntity)
		}

		if bindErr.Message != "name fields are required" {
			st.Errorf("expected error message to be %s; got %s", ErrBindNonPointer.Message, bindErr.Message)
		}

		person.Gender = ""

		bindErr = ValidateStruct(&person)

		if bindErr.HTTPStatusCode != http.StatusUnprocessableEntity {
			st.Errorf("expected HTTPStatusCode error to be %d; got %d", ErrBindNonPointer.HTTPStatusCode, http.StatusUnprocessableEntity)
		}

		if bindErr.Message != "name & gender fields are required" {
			st.Errorf("expected error message to be name & gender fields are required; got %s", bindErr.Message)
		}

		person.Email = ""

		bindErr = ValidateStruct(&person)

		if bindErr.HTTPStatusCode != http.StatusUnprocessableEntity {
			st.Errorf("expected HTTPStatusCode error to be %d; got %d", ErrBindNonPointer.HTTPStatusCode, http.StatusUnprocessableEntity)
		}

		if bindErr.Message != "name, gender, email fields are required" {
			st.Errorf("expected error message to be name, gender, email fields are required; got %s", bindErr.Message)
		}
	})

}

func TestNestedStructValidation(t *testing.T) {
	type Person struct {
		Name    string `form:"name" json:"name" rules:"required"`
		Gender  string `form:"gender" json:"gender" rules:"required"`
		Address struct {
			CityID     int `form:"city_id" json:"city_id" rules:"required"`
			PostalCode int `form:"psotal_code" json:"postal_code"`
		}
	}

	person := Person{
		Name:   "foo",
		Gender: "",
		Address: struct {
			CityID     int `form:"city_id" json:"city_id" rules:"required"`
			PostalCode int `form:"psotal_code" json:"postal_code"`
		}{
			CityID:     0,
			PostalCode: 204,
		},
	}

	errBind := ValidateStruct(&person)

	if errBind.HTTPStatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected error HTTPStatusCode to be %d; got %d", http.StatusUnprocessableEntity, errBind.HTTPStatusCode)
	}

	if errBind.Message != "gender & city_id fields are required" {
		t.Errorf("expected error message to be gender & city_id fields are required; got %s", errBind.Message)
	}
}
