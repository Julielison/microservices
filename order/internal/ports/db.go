package ports

import "github.com/Julielison/microservices/order/internal/application/core/domain"

type DBPort interface {
	Get(id string) (domain.Order, error)
	Save(*domain.Order) error
	// ProductExists verifica se um product_code existe na tabela de estoque.
	ProductExists(productCode string) (bool, error)
}
