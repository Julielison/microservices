package domain

import "time"

type ShippingItem struct {
	ProductCode string `json:"product_code"`
	Quantity    int32  `json:"quantity"`
}

type Shipping struct {
	ID               int64          `json:"id"`
	OrderID          int64          `json:"order_id"`
	ShippingItems    []ShippingItem `json:"shipping_items"`
	DeliveryDeadline int32          `json:"delivery_deadline"` // dias
	CreatedAt        int64          `json:"created_at"`
}

func NewShipping(orderId int64, items []ShippingItem) Shipping {
	return Shipping{
		CreatedAt:        time.Now().Unix(),
		OrderID:          orderId,
		ShippingItems:    items,
		DeliveryDeadline: CalculateDeadline(items),
	}
}

// CalculateDeadline calcula o prazo de entrega:
// mínimo 1 dia + 1 dia a cada 5 unidades totais.
func CalculateDeadline(items []ShippingItem) int32 {
	var total int32
	for _, item := range items {
		total += item.Quantity
	}
	// 1 dia base + floor(total / 5) dias extras
	return 1 + total/5
}
