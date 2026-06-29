package ports

import (
	"context"

	"github.com/Julielison/microservices/shipping/internal/application/core/domain"
)

type APIPort interface {
	Ship(ctx context.Context, shipping domain.Shipping) (domain.Shipping, error)
}
