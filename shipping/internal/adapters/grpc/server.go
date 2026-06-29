package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Julielison/microservices-proto/golang/shipping"
	"github.com/Julielison/microservices/shipping/config"
	"github.com/Julielison/microservices/shipping/internal/application/core/domain"
	"github.com/Julielison/microservices/shipping/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Adapter struct {
	api  ports.APIPort
	port int
	shipping.UnimplementedShippingServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{api: api, port: port}
}

func (a Adapter) Create(ctx context.Context, req *shipping.CreateShippingRequest) (*shipping.CreateShippingResponse, error) {
	log.Printf("Creating shipping for order %d...", req.OrderId)

	var items []domain.ShippingItem
	for _, it := range req.ShippingItems {
		items = append(items, domain.ShippingItem{
			ProductCode: it.ProductCode,
			Quantity:    it.Quantity,
		})
	}

	newShipping := domain.NewShipping(req.OrderId, items)
	result, err := a.api.Ship(ctx, newShipping)
	if err != nil {
		return nil, status.New(codes.Internal, fmt.Sprintf("failed to create shipping: %v", err)).Err()
	}

	return &shipping.CreateShippingResponse{
		ShippingId:       result.ID,
		DeliveryDeadline: result.DeliveryDeadline,
	}, nil
}

func (a Adapter) Run() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen on port %d, error: %v", a.port, err)
	}
	grpcServer := grpc.NewServer()
	shipping.RegisterShippingServer(grpcServer, a)
	if config.GetEnv() == "development" {
		reflection.Register(grpcServer)
	}
	log.Printf("starting shipping service on port %d ...", a.port)
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve grpc on port %d: %v", a.port, err)
	}
}
