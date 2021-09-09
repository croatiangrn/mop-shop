package mop_shop

import (
	"errors"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

type UserOrder struct {
	ID                      int       `gorm:"primaryKey;" json:"id"`
	UserID                  int       `gorm:"not null;index:ix_user_order_id;" json:"user_id"`
	TotalPrice              float32   `gorm:"not null;" json:"total_price"`
	StripeSessionID         *string   `gorm:"type:varchar(255);index:ix_stripe_session_id;" json:"stripe_session_id"`
	CreatedAt               time.Time `gorm:"not null;" json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	StripeClientReferenceID string    `gorm:"type:varchar(36);" json:"stripe_client_reference_id"`
	IsCompleted             bool      `gorm:"default:false;" json:"-"`
	orderItems              map[int]ItemWithStripeInfo
	db                      *gorm.DB
}

func (o *UserOrder) GetOrderItems() map[int]ItemWithStripeInfo {
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
	// Price is a virtual helper field
	Price float32
	// Quantity is a virtual field and is being used as quantity when creating stripe.CheckoutSessionLineItemParams
	Quantity int
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

func (o *UserOrder) CreateEmptyOrder(userID int, clientReferenceID string) error {
	query := `INSERT INTO user_orders (user_id, total_price, created_at, stripe_client_reference_id) VALUES (?, ?, ?, ?)`

	if err := o.db.Debug().Exec(query, userID, 0, time.Now(), clientReferenceID).Error; err != nil {
		log.Printf("error while creating empty order: %v\n", err)
		return ErrInternal
	}

	return nil
}

func (o *UserOrder) getProductsFromOrderBySessionID(sessionID string) (map[string]ItemWithStripeInfo, error) {
	i := session.ListLineItems(sessionID, nil)
	var productStripeIDs []string

	products := make(map[string]ItemWithStripeInfo)

	for i.Next() {
		li := i.LineItem()
		productStripeIDs = append(productStripeIDs, li.Price.Product.ID)
		products[li.Price.Product.ID] = ItemWithStripeInfo{
			UniqueStripePriceLookupKey: li.Price.LookupKey,
			StripeProductApiID:         li.Price.Product.ID,
			Price:                      float32(li.Price.UnitAmount),
			Quantity:                   int(li.Quantity),
		}
	}

	var dbShopItems []ItemWithStripeInfo
	query := `SELECT id AS item_id, unique_stripe_price_lookup_key, item_price, item_sale_price, stripe_product_api_id FROM shop_items WHERE stripe_product_api_id IN (?)`
	if err := o.db.Debug().Raw(query, productStripeIDs).Scan(&dbShopItems).Error; err != nil {
		log.Printf("error while getting shop items: %v\n", err)
		return nil, ErrInternal
	}

	for j := range dbShopItems {
		if obj, ok := products[dbShopItems[j].StripeProductApiID]; ok {
			obj.ItemID = dbShopItems[j].ItemID
			products[dbShopItems[j].StripeProductApiID] = obj
		}
	}

	return products, nil
}

func (o *UserOrder) UpdateEmptyOrderAfterCheckout(sessionID string, totalPrice float32) error {
	if o.ID == 0 {
		return ErrInvalidUserOrderID
	}

	o.TotalPrice = totalPrice

	products, err := o.getProductsFromOrderBySessionID(sessionID)
	if err != nil {
		return err
	}

	tx := o.db.Debug().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := `UPDATE user_orders SET updated_at = ?, is_completed = ?, total_price = ? WHERE stripe_session_id = ?`
	if err := tx.Debug().Exec(query, time.Now(), true, totalPrice, sessionID).Error; err != nil {
		tx.Rollback()
		log.Printf("error while updating user order: %v\n", err)
		return ErrInternal
	}

	var orderItemsQuerySB strings.Builder
	var orderItemsQueryParams []interface{}

	orderItemsQuery := `INSERT INTO user_order_items (user_order_id, shop_item_id, item_price, quantity) VALUES `
	orderItemsQuerySB.WriteString(orderItemsQuery)

	lastAvailableIndex := len(products) - 1
	counter := 0
	for i := range products {
		if counter == lastAvailableIndex {
			orderItemsQuerySB.WriteString(`(?, ?, ?, ?) `)
		} else {
			orderItemsQuerySB.WriteString(`(?, ?, ?, ?), `)
		}

		orderItemsQueryParams = append(orderItemsQueryParams, o.ID, products[i].Price, products[i].ItemID, products[i].Quantity)
		counter++
	}

	if err := tx.Debug().Exec(orderItemsQuerySB.String(), orderItemsQueryParams...).Error; err != nil {
		tx.Rollback()
		log.Printf("error while inserting user order items: %v\n", err)
		return ErrInternal
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("error while comitting transaction in userOrder.UpdateEmptyOrderAfterCheckout: %v\n", err)
		return ErrCommittingTransaction
	}

	return nil
}

func (o *UserOrder) FindOneByClientReferenceID(clientReferenceID string, orderCompleted bool) error {
	query := `SELECT * FROM user_orders WHERE stripe_client_reference_id = ? AND is_completed = ?`

	if err := o.db.Debug().Raw(query, clientReferenceID, orderCompleted).Take(o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		log.Printf("error while getting user order by client reference id: %v\n", err)
		return ErrInternal
	}

	return nil
}

func (o *UserOrder) PrepareForOrder(data *CreateUserOrder) error {
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

			obj.Quantity = data.Items[i].Quantity
			itemsWithStripeInfo[data.Items[i].ItemID] = obj

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
	ShopItemID  int     `gorm:"not null;" json:"shop_item_id"`
	ItemPrice   float32 `gorm:"not null;" json:"item_price"`
	Quantity    int     `gorm:"not null;"`
}

func (i *UserOrderItem) TableName() string {
	return "user_order_items"
}
