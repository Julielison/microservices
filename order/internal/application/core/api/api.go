package api

import (
	"github.com/ruandg/microservices/order/internal/application/core/domain"
	"github.com/ruandg/microservices/order/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db       ports.DBPort
	payment  ports.PaymentPort
	shipping ports.ShippingPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort, shipping ports.ShippingPort) *Application {
	return &Application{
		db:       db,
		payment:  payment,
		shipping: shipping,
	}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
	// 1. Validação: máximo de 50 itens no total
	var totalQuantity int32
	for _, item := range order.OrderItems {
		totalQuantity += item.Quantity
	}
	if totalQuantity > 50 {
		return domain.Order{}, status.Errorf(codes.InvalidArgument,
			"Order cannot have more than 50 items in total.")
	}

	// 2. Validação: verifica se todos os product_codes existem no estoque
	for _, item := range order.OrderItems {
		exists, err := a.db.ProductExists(item.ProductCode)
		if err != nil {
			return domain.Order{}, status.Errorf(codes.Internal,
				"failed to verify product %s: %v", item.ProductCode, err)
		}
		if !exists {
			return domain.Order{}, status.Errorf(codes.NotFound,
				"product %s does not exist in stock", item.ProductCode)
		}
	}

	// 3. Salva o pedido com status "Pending"
	err := a.db.Save(&order)
	if err != nil {
		return domain.Order{}, err
	}

	// 4. Solicita cobrança ao serviço Payment
	paymentErr := a.payment.Charge(&order)
	if paymentErr != nil {
		order.Status = "Canceled"
		_ = a.db.Save(&order)
		return domain.Order{}, paymentErr
	}

	// 5. Solicita envio ao serviço Shipping (somente após pagamento bem-sucedido)
	deadline, shippingErr := a.shipping.Ship(&order)
	if shippingErr != nil {
		order.Status = "Canceled"
		_ = a.db.Save(&order)
		return domain.Order{}, shippingErr
	}

	// 6. Atualiza status para "Paid" e salva o prazo de entrega
	order.Status = "Paid"
	order.DeliveryDeadline = deadline
	err = a.db.Save(&order)
	if err != nil {
		return domain.Order{}, err
	}

	return order, nil
}
