package example

import "ai-go-chi-starter/internal/service/shared"

const CodeNotFound = "EXAMPLE_NOT_FOUND"

func ErrNotFound() error {
	return shared.ErrNotFound(CodeNotFound, "example not found")
}
