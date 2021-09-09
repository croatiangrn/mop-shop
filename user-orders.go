package mop_shop

import (
	"gorm.io/gorm"
	"log"
	"strings"
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
	itemID                     int
	uniqueStripePriceLookupKey string
	itemPrice                  float32
	itemSalePrice              *float32
}

func findItemIDsWithStripePriceID(itemIDs []int, db *gorm.DB) (map[int]itemIDWithStripePriceID, error) {
	var data []itemIDWithStripePriceID
	query := `SELECT id AS item_id, stripe_price_api_id, item_price, item_sale_price FROM shop_items WHERE id IN (?)`

	if err := db.Debug().Raw(query, itemIDs).Scan(&data).Error; err != nil {
		log.Printf("error while getting findItemIDsWithStripePriceID: %v\n", err)
		return nil, ErrInternal
	}

	mapToReturn := make(map[int]itemIDWithStripePriceID, len(data))

	for i := range data {
		mapToReturn[data[i].itemID] = data[i]
	}

	return mapToReturn, nil
}

func (o *UserOrder) Create(data *CreateUserOrder) error {
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

	orderTotalPriceAmount := float32(0)

	for i := range data.Items {
		if obj, ok := itemIDsWithStripePriceIDs[data.Items[i].ItemID]; ok {
			price := obj.itemPrice
			if obj.itemSalePrice != nil {
				price = *obj.itemSalePrice
			}

			data.Items[i].itemPrice = price
			orderTotalPriceAmount += price
		}
	}

	tx := o.db.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	insertOrderQuery := `INSERT INTO user_orders (user_id, total_price, created_at) VALUES (?, ?, ?)`
	if err := tx.Debug().Exec(insertOrderQuery, o.UserID, orderTotalPriceAmount, time.Now()).Error; err != nil {
		tx.Rollback()
		log.Printf("error while inserting user order: %v\n", err)
		return ErrInternal
	}

	orderID, err := getLastInsertedID(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	var orderItemsQuerySB strings.Builder
	var orderItemsQueryParams []interface{}

	orderItemsQuery := `INSERT INTO user_order_items (user_order_id, item_price, quantity) VALUES `
	orderItemsQuerySB.WriteString(orderItemsQuery)

	lastAvailableIndex := len(data.Items) - 1
	for i := range data.Items {
		if i == lastAvailableIndex {
			orderItemsQuerySB.WriteString(`(?, ?, ?) `)
		} else {
			orderItemsQuerySB.WriteString(`(?, ?, ?), `)
		}

		orderItemsQueryParams = append(orderItemsQueryParams, orderID, data.Items[i].itemPrice, data.Items[i].Quantity)
	}

	if err := tx.Debug().Exec(orderItemsQuerySB.String(), orderItemsQueryParams...).Error; err != nil {
		tx.Rollback()
		log.Printf("error while inserting user order items: %v\n", err)
		return ErrInternal
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("error while committing transaction: %v\n", err)
		return ErrInternal
	}

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
