package shared

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationDetails struct {
	FieldErrors []FieldError `json:"field_errors,omitempty"`
}

func RequiredField(field string) FieldError {
	return FieldError{
		Field:   field,
		Message: "is required",
	}
}

func WithFieldErrors(fieldErrors ...FieldError) ErrorOption {
	filtered := make([]FieldError, 0, len(fieldErrors))
	for _, fieldError := range fieldErrors {
		if fieldError.Field == "" && fieldError.Message == "" {
			continue
		}
		filtered = append(filtered, fieldError)
	}
	return WithDetails(ValidationDetails{FieldErrors: filtered})
}
