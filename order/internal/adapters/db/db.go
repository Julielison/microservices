package db

import (
	"fmt"

	"github.com/Julielison/microservices/order/internal/application/core/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// StockItem representa um produto disponível no estoque.
// Deve ser populado manualmente ou via seed antes de aceitar pedidos.
type StockItem struct {
	gorm.Model
	ProductCode string `gorm:"uniqueIndex;not null"`
	Description string
	UnitPrice   float32
}

type Order struct {
	gorm.Model
	CustomerID       int64
	Status           string
	DeliveryDeadline int32
	OrderItems       []OrderItem
}

type OrderItem struct {
	gorm.Model
	ProductCode string
	UnitPrice   float32
	Quantity    int32
	OrderID     uint
}

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(dataSourceUrl string) (*Adapter, error) {
	db, openErr := gorm.Open(mysql.Open(dataSourceUrl), &gorm.Config{})
	if openErr != nil {
		return nil, fmt.Errorf("db connection error: %v", openErr)
	}
	err := db.AutoMigrate(&Order{}, &OrderItem{}, &StockItem{})
	if err != nil {
		return nil, fmt.Errorf("db migration error: %v", err)
	}
	return &Adapter{db: db}, nil
}

func (a Adapter) Get(id string) (domain.Order, error) {
	var orderEntity Order
	res := a.db.Preload("OrderItems").First(&orderEntity, id)
	var orderItems []domain.OrderItem
	for _, orderItem := range orderEntity.OrderItems {
		orderItems = append(orderItems, domain.OrderItem{
			ProductCode: orderItem.ProductCode,
			UnitPrice:   orderItem.UnitPrice,
			Quantity:    orderItem.Quantity,
		})
	}
	order := domain.Order{
		ID:               int64(orderEntity.ID),
		CustomerID:       orderEntity.CustomerID,
		Status:           orderEntity.Status,
		OrderItems:       orderItems,
		DeliveryDeadline: orderEntity.DeliveryDeadline,
		CreatedAt:        orderEntity.CreatedAt.UnixNano(),
	}
	return order, res.Error
}

func (a Adapter) Save(order *domain.Order) error {
	// Se já tem ID, atualiza status e delivery_deadline
	if order.ID != 0 {
		res := a.db.Model(&Order{}).Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"status":            order.Status,
				"delivery_deadline": order.DeliveryDeadline,
			})
		return res.Error
	}

	// Novo pedido
	var orderItems []OrderItem
	for _, item := range order.OrderItems {
		orderItems = append(orderItems, OrderItem{
			ProductCode: item.ProductCode,
			UnitPrice:   item.UnitPrice,
			Quantity:    item.Quantity,
		})
	}
	orderModel := Order{
		CustomerID: order.CustomerID,
		Status:     order.Status,
		OrderItems: orderItems,
	}
	res := a.db.Create(&orderModel)
	if res.Error == nil {
		order.ID = int64(orderModel.ID)
	}
	return res.Error
}

// ProductExists verifica se um product_code está cadastrado na tabela de estoque.
func (a Adapter) ProductExists(productCode string) (bool, error) {
	var count int64
	res := a.db.Model(&StockItem{}).
		Where("product_code = ?", productCode).
		Count(&count)
	return count > 0, res.Error
}
