package api

import (
	"github.com/ruandg/microservices/order/internal/application/core/domain"
	"github.com/ruandg/microservices/order/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db      ports.DBPort
	payment ports.PaymentPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort) *Application {
	return &Application{
		db:      db,
		payment: payment,
	}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
	// Validação: máximo de 50 itens no total
	var totalQuantity int32
	for _, item := range order.OrderItems {
		totalQuantity += item.Quantity
	}
	if totalQuantity > 50 {
		return domain.Order{}, status.Errorf(codes.InvalidArgument, "Order cannot have more than 50 items in total.")
	}

	// Salva o pedido (status inicial "Pending", definido em domain.NewOrder)
	err := a.db.Save(&order)
	if err != nil {
		return domain.Order{}, err
	}

	// Solicita cobrança ao serviço Payment
	paymentErr := a.payment.Charge(&order)
	if paymentErr != nil {
		// Atualiza status para "Canceled" no banco
		order.Status = "Canceled"
		_ = a.db.Save(&order)
		return domain.Order{}, paymentErr
	}

	// Atualiza status para "Paid"
	order.Status = "Paid"
	err = a.db.Save(&order)
	if err != nil {
		return domain.Order{}, err
	}

	return order, nil
}
