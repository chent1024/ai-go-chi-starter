package example

import "context"

type Repository interface {
	Create(ctx context.Context, item Example) (Example, error)
	Get(ctx context.Context, id string) (Example, error)
	List(ctx context.Context) ([]Example, error)
}
