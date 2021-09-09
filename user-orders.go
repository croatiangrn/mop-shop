package mop_shop

import (
	"gorm.io/gorm"
	"log"
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

type itemIDWithStripePriceID struct {
	itemID           int
	stripePriceApiID string
}

func findItemIDsWithStripePriceID(itemIDs []int, db *gorm.DB) ([]itemIDWithStripePriceID, error) {
	var data []itemIDWithStripePriceID
	query := `SELECT id AS item_id, stripe_price_api_id FROM shop_items WHERE id IN (?)`

	if err := db.Debug().Raw(query, itemIDs).Scan(&data).Error; err != nil {
		log.Printf("error while getting findItemIDsWithStripePriceID: %v\n", err)
		return nil, ErrInternal
	}

	return data, nil
}

func (o *UserOrder) Create(data *CreateUserOrder) error {
	if data == nil {
		return ErrOrderDataBlank
	}

	if err := data.Validate(); err != nil {
		return err
	}

	var itemIDs []int
	for i := range data.Items {
		itemIDs = append(itemIDs, data.Items[i].ItemID)
	}

	// TODO: Fetch price IDs from provided data.Items

	// TODO: Insert DB queries here and create checkout!

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
