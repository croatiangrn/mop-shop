package mop_shop

import (
	"github.com/davecgh/go-spew/spew"
	"gorm.io/gorm"
	"log"
	"time"
)

type UserOrder struct {
	ID              int       `gorm:"primaryKey;" json:"id"`
	UserID          int       `gorm:"not null;index:ix_user_order_id;" json:"user_id"`
	TotalPrice      float32   `gorm:"not null;" json:"total_price"`
	StripeSessionID *string   `gorm:"type:varchar(255);" json:"stripe_session_id"`
	CreatedAt       time.Time `gorm:"not null;" json:"created_at"`
	db              *gorm.DB
}

func (o *UserOrder) TableName() string {
	return "user_orders"
}

func NewUserOrder(db *gorm.DB) *UserOrder {
	return &UserOrder{db: db}
}

type itemIDWithStripePriceID struct {
	ItemID                     int
	UniqueStripePriceLookupKey string
	ItemPrice                  float32
	ItemSalePrice              *float32
}

func findItemIDsWithStripePriceID(itemIDs []int, db *gorm.DB) (map[int]itemIDWithStripePriceID, error) {
	var data []itemIDWithStripePriceID
	query := `SELECT id AS item_id, unique_stripe_price_lookup_key, item_price, item_sale_price FROM shop_items WHERE id IN (?)`

	if err := db.Debug().Raw(query, itemIDs).Scan(&data).Error; err != nil {
		log.Printf("error while getting findItemIDsWithStripePriceID: %v\n", err)
		return nil, ErrInternal
	}

	spew.Dump(data)

	mapToReturn := make(map[int]itemIDWithStripePriceID, len(data))

	for i := range data {
		mapToReturn[data[i].ItemID] = data[i]
	}

	return mapToReturn, nil
}

func (o *UserOrder) Create(data *CreateUserOrder) error {
	o.UserID = data.userID

	if data == nil {
		return ErrOrderDataBlank
	}

	if err := data.validate(); err != nil {
		return err
	}

	var itemIDs []int
	for i := range data.Items {
		itemIDs = append(itemIDs, data.Items[i].ItemID)
	}

	itemIDsWithStripePriceIDs, err := findItemIDsWithStripePriceID(itemIDs, o.db)
	if err != nil {
		return err
	}

	if len(itemIDsWithStripePriceIDs) != len(data.Items) {
		return ErrSomeItemsDoNotExist
	}

	orderTotalPriceAmount := float32(0)

	for i := range data.Items {
		if obj, ok := itemIDsWithStripePriceIDs[data.Items[i].ItemID]; ok {
			price := obj.ItemPrice
			if obj.ItemSalePrice != nil {
				price = *obj.ItemSalePrice
			}

			data.Items[i].itemPrice = price
			orderTotalPriceAmount += price
		}
	}

	o.TotalPrice = orderTotalPriceAmount
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
