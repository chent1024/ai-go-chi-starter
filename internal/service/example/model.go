package example

import "time"

type Example struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateInput struct {
	Name string
}
