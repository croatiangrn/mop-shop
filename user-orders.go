package mop_shop

import (
	"encoding/json"
	"errors"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"gorm.io/gorm"
	"log"
	"net/url"
	"strconv"
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

func (o *UserOrder) TableName() string {
	return "user_orders"
}

func (o *UserOrder) GetOrderItems() map[int]ItemWithStripeInfo {
	return o.orderItems
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

func (o *UserOrder) UpdateEmptyOrderAfterCheckout(sessionID, clientReferenceID string, totalPrice float32) error {
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

	query := `UPDATE user_orders SET updated_at = ?, is_completed = ?, total_price = ?, stripe_session_id = ? WHERE stripe_client_reference_id = ?`

	orderQuery := tx.Debug().Exec(query, time.Now(), true, totalPrice, sessionID, clientReferenceID)

	if err := orderQuery.Error; err != nil {
		tx.Rollback()
		log.Printf("error while updating user order: %v\n", err)
		return ErrInternal
	}

	if orderQuery.RowsAffected == 0 {
		tx.Rollback()
		log.Printf("user order with client reference id %q not found\n", clientReferenceID)
		return gorm.ErrRecordNotFound
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

		orderItemsQueryParams = append(orderItemsQueryParams, o.ID, products[i].ItemID, products[i].Price, products[i].Quantity)
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

// UserOrderFrontResponse struct should be used for user requests such as getting all user orders, single user order
type UserOrderFrontResponse struct {
	ID          int       `gorm:"primaryKey;" json:"id"`
	TotalPrice  float32   `gorm:"not null;" json:"total_price"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `gorm:"not null;" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsCompleted bool      `gorm:"default:false;" json:"-"`
	// Items will not be shown in JSON response if it's nil!
	RawItems json.RawMessage              `json:"raw_items,omitempty"`
	Items    []UserOrderItemFrontResponse `gorm:"-" json:"items,omitempty"` // GORM -> Ignore this field as it will be manually unmarshalled
	db       *gorm.DB
}

type UserOrderItemFrontResponse struct {
	ItemID          int     `json:"item_id"`
	ItemName        string  `json:"item_name"`
	ItemPrice       float32 `json:"item_price"`
	ItemPicture     string  `json:"item_picture"`
	ItemDescription string  `json:"item_description"`
	Quantity        int     `json:"quantity"`
}

func GetUserOrders(userID int, db *gorm.DB, paginationParams PaginationParams, currency, currentURL string) ([]UserOrderFrontResponse, *PaginationResponse, error) {
	paginationParams.normalize()

	query := `SELECT 
			uo.id, uo.total_price, uo.created_at, uo.updated_at, uo.is_completed, json_arrayagg(
				json_object(
					'item_id', si.id,
					'item_name', si.item_name,
					'item_price', uoi.item_price,
					'item_picture', si.item_picture,
					'item_description', si.item_description,
					'quantity', uoi.quantity
				)
			) AS raw_items 
		FROM user_orders uo
		INNER JOIN users u ON u.id = uo.user_id AND u.deleted_at IS NULL
		INNER JOIN user_order_items uoi ON uoi.user_order_id = uo.id
		INNER JOIN shop_items si ON si.id = uoi.shop_item_id
		WHERE uo.user_id = ? `

	var userOrdersQuery strings.Builder
	var params []interface{}

	userOrdersQuery.WriteString(query)

	switch {
	case paginationParams.Before == 0 && paginationParams.After == 0:
		userOrdersQuery.WriteString(`AND uo.id > ? GROUP BY uo.id ORDER BY uo.id DESC LIMIT ?`)
		params = append(params, 0)
	case paginationParams.After > 0:
		userOrdersQuery.WriteString(`AND uo.id <= ? GROUP BY uo.id ORDER BY uo.id DESC LIMIT ?`)
		params = append(params, paginationParams.After)
	case paginationParams.Before > 0:
		userOrdersQuery.WriteString(`AND uo.id >= ? GROUP BY uo.id ORDER BY uo.id ASC LIMIT ?`)
		params = append(params, paginationParams.Before)
	}

	params = append(params, paginationParams.PerPage+1)

	var data []UserOrderFrontResponse
	if err := db.Debug().Raw(query, userID).Scan(&data).Error; err != nil {
		log.Printf("error while getting user orders: %v\n", err)
		return nil, nil, ErrInternal
	}

	for i := range data {
		data[i].Currency = currency

		if err := json.Unmarshal(data[i].RawItems, &data[i].Items); err != nil {
			log.Printf("error while unmarshalling user order items: %v\n", err)
			return nil, nil, ErrInternal
		}

		data[i].RawItems = nil
	}

	if len(data) == 0 {
		data = []UserOrderFrontResponse{}
	}

	pages := PaginationResponse{}

	parsedURL, err := url.Parse(currentURL)
	if err != nil {
		return nil, nil, ErrParsingURL
	}

	beforeQ := parsedURL.Query()
	afterQ := parsedURL.Query()

	switch {
	case len(data) == paginationParams.PerPage+1:

		if paginationParams.After > 0 {
			beforeId := strconv.Itoa(data[0].ID + 1)

			beforeQ.Set("before", beforeId)
			parsedURL.RawQuery = beforeQ.Encode()
			cursorBefore := parsedURL.String()

			pages.CursorBefore = &cursorBefore
			pages.Before = &beforeId
		} else if paginationParams.Before > 0 {
			beforeId := strconv.Itoa(data[0].ID)

			beforeQ.Set("before", beforeId)
			parsedURL.RawQuery = beforeQ.Encode()
			cursorBefore := parsedURL.String()

			pages.Before = &beforeId
			pages.CursorBefore = &cursorBefore
		}

		afterId := strconv.Itoa(data[len(data)-1].ID)

		afterQ.Set("after", afterId)
		parsedURL.RawQuery = afterQ.Encode()
		cursorAfter := parsedURL.String()

		pages.CursorAfter = &cursorAfter
		pages.After = &afterId

		if paginationParams.Before > 0 {
			afterId := strconv.Itoa(data[len(data)-1].ID - 1)

			afterQ.Set("after", afterId)
			parsedURL.RawQuery = afterQ.Encode()
			cursorAfter := parsedURL.String()

			pages.After = &afterId
			pages.CursorAfter = &cursorAfter
			data = data[1:]
		} else {
			data = data[:len(data)-1]
		}
	case len(data) <= paginationParams.PerPage && paginationParams.Before > 0:
		afterId := strconv.Itoa(data[len(data)-1].ID - 1)

		afterQ.Set("after", afterId)
		parsedURL.RawQuery = afterQ.Encode()
		cursorAfter := parsedURL.String()

		pages.After = &afterId
		pages.CursorAfter = &cursorAfter
	case len(data) <= paginationParams.PerPage && paginationParams.After > 0:
		beforeId := strconv.Itoa(data[0].ID + 1)

		beforeQ.Set("before", beforeId)
		parsedURL.RawQuery = beforeQ.Encode()
		cursorBefore := parsedURL.String()

		pages.Before = &beforeId
		pages.CursorBefore = &cursorBefore
	}

	return data, &pages, nil
}
