package grpc

import (
	"context"
	"fmt"

	"github.com/huseyinbabal/microservices/payment/internal/application/core/domain"
	"github.com/ruandg/microservices-proto/golang/payment"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a Adapter) Create(ctx context.Context, request *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	log.WithContext(ctx).Info("Creating payment...")

	newPayment := domain.NewPayment(request.UserId, request.OrderId, request.TotalPrice)
	result, err := a.api.Charge(ctx, newPayment)
	code := status.Code(err)
	if code == codes.InvalidArgument {
		return nil, err
	} else if err != nil {
		return nil, status.New(codes.Internal, fmt.Sprintf("failed to charge. %v ", err)).Err()
	}
	return &payment.CreatePaymentResponse{PaymentId: result.ID}, nil
}
