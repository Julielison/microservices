package payment_adapter

import (
	"context"
	"log"
	"time"

	"github.com/Julielison/microservices-proto/golang/payment"
	"github.com/Julielison/microservices/order/internal/application/core/domain"
	grpc_retry "github.com/g
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Adapter struct {
	payment payment.PaymentClient
}

func NewAdapter(paymentServiceUrl string) (*Adapter, error) {
	var opts []grpc.DialOption

	// Interceptor de retry: até 5 tentativas nos códigos Unavailable e ResourceExhausted,
	// com backoff linear de 1 segundo incrementado a cada tentativa.
	opts = append(opts,
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
			grpc_retry.WithMax(5),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
		)),
	)

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(paymentServiceUrl, opts...)
	if err != nil {
		return nil, err
	}

	client := payment.NewPaymentClient(conn)
	return &Adapter{payment: client}, nil
}

func (a *Adapter) Charge(order *domain.Order) error {
	// Deadline individual de 2 segundos por chamada ao serviço Payment.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := a.payment.Create(ctx, &payment.CreatePaymentRequest{
		UserId:     order.CustomerID,
		OrderId:    order.ID,
		TotalPrice: order.TotalPrice(),
	})

	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			log.Printf("[Payment] Timeout (DeadlineExceeded) ao cobrar o pedido %d do cliente %d: %v",
				order.ID, order.CustomerID, err)
		}
		return err
	}

	return nil
}
