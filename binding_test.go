package nano

import (
	"bytes"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAutoBindingForUnexpectedContentType(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/", nil)
	if err != nil {
		log.Fatalf("could not create http request: %v", err)
	}
	req.Header.Add(HeaderContentType, "x-unknown")
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	type Person struct {
		Name   string `form:"name"`
		Gender string `form:"gender"`
	}

	var person Person
	if err = ctx.Bind(&person); err == nil {
		t.Fatalf("expected error returned")
	}

	if err, ok := err.(ErrBinding); ok {
		if err.Status != ErrBindContentType.Status {
			t.Errorf("expected error HTTPStatusCode to be %d; got %d", ErrBindContentType.Status, err.Status)
		}

		if err.Text != ErrBindContentType.Text {
			t.Errorf("expected error message to be %s; got %s", ErrBindContentType.Text, err.Text)
		}

		return
	}

	t.Fatalf("expected ErrBinding type returned, got %T", err)

}

func TestAutoBindingForURLEncoded(t *testing.T) {
	form := url.Values{}
	form.Set("name", "foo")
	form.Set("gender", "male")

	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatalf("could not create http request: %v", err)
	}
	req.Header.Add(HeaderContentType, MimeFormURLEncoded)
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	type Person struct {
		Name   string `form:"name"`
		Gender string `form:"gender"`
	}

	var person Person
	errBinding := ctx.Bind(&person)

	if nm := ctx.PostForm("name"); nm != "foo" {
		t.Fatalf("expected form name value to be foo; got %s", nm)
	}

	if errBinding != nil {
		t.Fatalf("expected err binding to nil")
	}

	if person.Name != "foo" {
		t.Errorf("expected name to be foo; got %s", person.Name)
	}

	if person.Gender != "male" {
		t.Errorf("expected gender to be male; got %s", person.Gender)
	}
}

func TestAutoBindingForJSON(t *testing.T) {
	form := []byte(`{"name":"foo", "gender":"male"}`)
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(form))
	if err != nil {
		log.Fatalf("could not create http request: %v", err)
	}
	req.Header.Add(HeaderContentType, MimeJSON)
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	type Person struct {
		Name   string `form:"name" json:"name"`
		Gender string `form:"gender" json:"gender"`
	}

	var person Person
	errBinding := ctx.Bind(&person)

	if errBinding != nil {
		t.Fatalf("expected err binding to nil")
	}

	if person.Name != "foo" {
		t.Errorf("expected name to be foo; got %s", person.Name)
	}

	if person.Gender != "male" {
		t.Errorf("expected gender to be male; got %s", person.Gender)
	}
}

func TestAutoBindingForMultipartForm(t *testing.T) {
	body := new(bytes.Buffer)
	form := multipart.NewWriter(body)
	form.WriteField("name", "foo")
	form.WriteField("gender", "male")

	req, err := http.NewRequest(http.MethodPost, "/", body)
	if err != nil {
		log.Fatalf("could not create http request: %v", err)
	}
	req.Header.Add(HeaderContentType, form.FormDataContentType())
	form.Close()
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	type Person struct {
		Name   string `form:"name" json:"name"`
		Gender string `form:"gender" json:"gender"`
	}

	var person Person

	if err = ctx.Bind(&person); err != nil {
		t.Fatalf("expected err binding to nil; got %T", err)
	}

	if person.Name != "foo" {
		t.Errorf("expected name to be foo; got %s", person.Name)
	}

	if person.Gender != "male" {
		t.Errorf("expected gender to be male; got %s", person.Gender)
	}
}

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
		w := httptest.NewRecorder()

		ctx := newContext(w, req)

		err = ctx.BindJSON(person)
		if err == nil {
			st.Errorf("expected error to be returned; got %T", err)
		}

		if errBinding, ok := err.(ErrBinding); ok {
			if errBinding.Error() != ErrBindNonPointer.Error() {
				st.Errorf("expect error to be ErrBindNonPointer; got %v", errBinding)
			}

			return
		}

		st.Fatalf("expected ErrBinding, got %T", err)

	})
}
