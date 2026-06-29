package ports

import (
	"context"

	"github.com/ruandg/microservices/shipping/internal/application/core/domain"
)

type DBPort interface {
	Save(ctx context.Context, shipping *domain.Shipping) error
	Get(ctx context.Context, id string) (domain.Shipping, error)
}
