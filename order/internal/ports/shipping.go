package ports

import "github.com/Julielison/microservices/order/internal/application/core/domain"

// ShippingPort define a interface que o núcleo da aplicação usa
// para solicitar o envio de um pedido ao microsserviço Shipping.
type ShippingPort interface {
	Ship(order *domain.Order) (int32, error) // retorna delivery_deadline em dias
}
