package utils

import "testing"

type sampleValidateRequest struct {
	Name  string  `json:"name" validate:"required"`
	Email string  `json:"email" validate:"required,email"`
	Price float64 `json:"price" validate:"required,gt=0"`
}

func TestValidateStruct_Valid(t *testing.T) {
	v := NewValidator()

	errs := v.ValidateStruct(sampleValidateRequest{
		Name:  "ok",
		Email: "a@b.com",
		Price: 10,
	})

	if errs != nil {
		t.Errorf("ValidateStruct() = %v, want nil", errs)
	}
}

func TestValidateStruct_Invalid(t *testing.T) {
	v := NewValidator()

	errs := v.ValidateStruct(sampleValidateRequest{
		Email: "not-an-email",
		Price: -5,
	})

	if errs == nil {
		t.Fatal("ValidateStruct() = nil, want validation errors")
	}

	// Reports failures keyed by the struct's json tag, not the Go field name.
	for _, field := range []string{"name", "email", "price"} {
		if len(errs[field]) == 0 {
			t.Errorf("expected at least one message for field %q, got none (errs=%v)", field, errs)
		}
	}
}

func TestValidateStruct_NotAStruct(t *testing.T) {
	v := NewValidator()

	// A non-struct value can't be validated by struct tags; ValidateStruct
	// should report it as valid (nil) rather than panic or misreport.
	errs := v.ValidateStruct(42)

	if errs != nil {
		t.Errorf("ValidateStruct(42) = %v, want nil", errs)
	}
}
