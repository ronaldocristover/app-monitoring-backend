package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Email string `validate:"required,email"`
	Name  string `validate:"required,min=2,max=50"`
	Age   int    `validate:"required,min=1"`
}

func TestStruct_Valid(t *testing.T) {
	s := testStruct{
		Email: "user@example.com",
		Name:  "John",
		Age:   25,
	}

	err := Struct(s)
	assert.NoError(t, err)
}

func TestStruct_MissingRequired(t *testing.T) {
	s := testStruct{
		Email: "",
		Name:  "John",
		Age:   25,
	}

	err := Struct(s)
	assert.Error(t, err)
}

func TestStruct_InvalidEmail(t *testing.T) {
	s := testStruct{
		Email: "not-an-email",
		Name:  "John",
		Age:   25,
	}

	err := Struct(s)
	assert.Error(t, err)
}

func TestStruct_MinViolation(t *testing.T) {
	s := testStruct{
		Email: "user@example.com",
		Name:  "J",
		Age:   25,
	}

	err := Struct(s)
	assert.Error(t, err)
}

func TestErrors_Nil(t *testing.T) {
	result := Errors(nil)
	assert.Nil(t, result)
}

func TestErrors_ReturnsMessages(t *testing.T) {
	s := testStruct{
		Email: "",
		Name:  "",
		Age:   0,
	}

	err := Struct(s)
	assert.Error(t, err)

	errors := Errors(err)
	assert.NotEmpty(t, errors)
	assert.Len(t, errors, 3)
}

func TestErrors_RequiredMessage(t *testing.T) {
	type req struct {
		Name string `validate:"required"`
	}

	err := Struct(req{Name: ""})
	errors := Errors(err)

	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "Name")
	assert.Contains(t, errors[0], "required")
}

func TestErrors_EmailMessage(t *testing.T) {
	type req struct {
		Email string `validate:"required,email"`
	}

	err := Struct(req{Email: "bad"})
	errors := Errors(err)

	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "Email")
	assert.Contains(t, errors[0], "valid email")
}

func TestErrors_MinMessage(t *testing.T) {
	type req struct {
		Name string `validate:"min=5"`
	}

	err := Struct(req{Name: "ab"})
	errors := Errors(err)

	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "Name")
	assert.Contains(t, errors[0], "at least 5")
}

func TestErrors_MaxMessage(t *testing.T) {
	type req struct {
		Name string `validate:"max=3"`
	}

	err := Struct(req{Name: "toolong"})
	errors := Errors(err)

	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "Name")
	assert.Contains(t, errors[0], "at most 3")
}

func TestField_Valid(t *testing.T) {
	err := Field("test@example.com", "required,email")
	assert.NoError(t, err)
}

func TestField_Invalid(t *testing.T) {
	err := Field("not-email", "email")
	assert.Error(t, err)
}
