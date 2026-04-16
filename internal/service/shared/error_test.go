package shared

import (
	"testing"
)

func TestErrInvalidArgumentCarriesFieldErrors(t *testing.T) {
	err := ErrInvalidArgument("name is required", WithFieldErrors(RequiredField("name")))

	if Code(err) != CodeInvalidArgument || HTTPStatus(err) != StatusBadRequest {
		t.Fatalf("unexpected error metadata: code=%q status=%d", Code(err), HTTPStatus(err))
	}

	details, ok := Details(err).(ValidationDetails)
	if !ok {
		t.Fatalf("Details() type = %T", Details(err))
	}
	if len(details.FieldErrors) != 1 {
		t.Fatalf("field error count = %d", len(details.FieldErrors))
	}
	if details.FieldErrors[0].Field != "name" || details.FieldErrors[0].Message != "is required" {
		t.Fatalf("unexpected field error = %+v", details.FieldErrors[0])
	}
}

func TestMarkRetryablePreservesDetails(t *testing.T) {
	err := MarkRetryable(ErrInvalidArgument("name is required", WithFieldErrors(RequiredField("name"))))

	if !Retryable(err) {
		t.Fatal("Retryable() = false")
	}

	details, ok := Details(err).(ValidationDetails)
	if !ok || len(details.FieldErrors) != 1 {
		t.Fatalf("Details() = %#v", Details(err))
	}
}
