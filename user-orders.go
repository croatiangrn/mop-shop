package mop_shop

import (
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
	orderItems      map[int]ItemWithStripeInfo
	db              *gorm.DB
}

func (o *UserOrder) OrderItems() map[int]ItemWithStripeInfo {
	return o.orderItems
}

func (o *UserOrder) TableName() string {
	return "user_orders"
}

func NewUserOrder(db *gorm.DB) *UserOrder {
	return &UserOrder{db: db}
}

type ItemWithStripeInfo struct {
	ItemID                     int
	UniqueStripePriceLookupKey string
	ItemPrice                  float32
	ItemSalePrice              *float32
	StripeProductApiID         string
	// Price is a virtual field and is being used as unit_amount when creating stripe checkout session
	Price float32
}

func findItemsWithStripeInfo(itemIDs []int, db *gorm.DB) (map[int]ItemWithStripeInfo, error) {
	var data []ItemWithStripeInfo
	query := `SELECT id AS item_id, stripe_product_api_id, unique_stripe_price_lookup_key, item_price, item_sale_price FROM shop_items WHERE id IN (?)`

	if err := db.Debug().Raw(query, itemIDs).Scan(&data).Error; err != nil {
		log.Printf("error while getting findItemsWithStripeInfo: %v\n", err)
		return nil, ErrInternal
	}

	mapToReturn := make(map[int]ItemWithStripeInfo, len(data))

	for i := range data {
		data[i].Price = data[i].ItemPrice

		if data[i].ItemSalePrice != nil {
			data[i].Price = *data[i].ItemSalePrice
		}

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

	itemsWithStripeInfo, err := findItemsWithStripeInfo(itemIDs, o.db)
	if err != nil {
		return err
	}

	if len(itemsWithStripeInfo) != len(data.Items) {
		return ErrSomeItemsDoNotExist
	}

	orderTotalPriceAmount := float32(0)

	for i := range data.Items {
		if obj, ok := itemsWithStripeInfo[data.Items[i].ItemID]; ok {
			price := obj.ItemPrice
			if obj.ItemSalePrice != nil {
				price = *obj.ItemSalePrice
			}

			data.Items[i].itemPrice = price
			orderTotalPriceAmount += price
		}
	}

	o.TotalPrice = orderTotalPriceAmount
	o.orderItems = itemsWithStripeInfo

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
