package mop_shop

import (
	"gorm.io/gorm"
	"time"
)

type UserOrder struct {
	ID         int       `gorm:"primaryKey;" json:"id"`
	UserID     int       `gorm:"not null;index:ix_user_order_id;" json:"user_id"`
	TotalPrice float32   `gorm:"not null;" json:"total_price"`
	CreatedAt  time.Time `gorm:"not null;" json:"created_at"`
	db         *gorm.DB
}

func (o *UserOrder) TableName() string {
	return "user_orders"
}

func NewUserOrder(db *gorm.DB) *UserOrder {
	return &UserOrder{db: db}
}

func (o *UserOrder) Create(items []UserOrder) error {
	return nil
}

func (o *UserOrder) GetOrderByID(orderID int) error {
	return nil
}

type UserOrderItem struct {
	ID          int     `gorm:"primaryKey;" json:"id"`
	UserOrderID int     `gorm:"not null;index:ix_user_order_item_order_id;" json:"user_order_id"`
	ItemPrice   float32 `gorm:"not null;" json:"item_price"`
	Quantity    int     `gorm:"not null;"`
}

func (i *UserOrderItem) TableName() string {
	return "user_order_items"
}
