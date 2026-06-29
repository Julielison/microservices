package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Julielison/microservices/shipping/internal/application/core/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Shipping é a entidade persistida no banco.
type Shipping struct {
	gorm.Model
	OrderID           int64  `gorm:"not null"`
	ShippingItemsJSON string `gorm:"type:text"`
	DeliveryDeadline  int32
}

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(dataSourceUrl string) (*Adapter, error) {
	db, err := gorm.Open(mysql.Open(dataSourceUrl), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("db connection error: %v", err)
	}
	if err := db.AutoMigrate(&Shipping{}); err != nil {
		return nil, fmt.Errorf("db migration error: %v", err)
	}
	return &Adapter{db: db}, nil
}

func (a Adapter) Save(ctx context.Context, s *domain.Shipping) error {
	raw, err := json.Marshal(s.ShippingItems)
	if err != nil {
		return err
	}
	entity := Shipping{
		OrderID:           s.OrderID,
		ShippingItemsJSON: string(raw),
		DeliveryDeadline:  s.DeliveryDeadline,
	}
	res := a.db.WithContext(ctx).Create(&entity)
	if res.Error == nil {
		s.ID = int64(entity.ID)
	}
	return res.Error
}

func (a Adapter) Get(ctx context.Context, id string) (domain.Shipping, error) {
	var entity Shipping
	res := a.db.WithContext(ctx).First(&entity, id)
	if res.Error != nil {
		return domain.Shipping{}, res.Error
	}
	var items []domain.ShippingItem
	_ = json.Unmarshal([]byte(entity.ShippingItemsJSON), &items)
	return domain.Shipping{
		ID:               int64(entity.ID),
		OrderID:          entity.OrderID,
		ShippingItems:    items,
		DeliveryDeadline: entity.DeliveryDeadline,
		CreatedAt:        entity.CreatedAt.UnixNano(),
	}, nil
}
