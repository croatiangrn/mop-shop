package mop_shop

import (
	"github.com/stripe/stripe-go/v72"
	"gorm.io/gorm"
	"log"
	"time"
)

type ShopItem struct {
	ID                 int        `gorm:"primaryKey" json:"id"`
	ItemName           string     `gorm:"not null;type:varchar(255);" json:"item_name"`
	ItemPicture        *string    `gorm:"default:null;type:varchar(255);" json:"item_picture"`
	ItemPrice          int64      `gorm:"not null;" json:"item_price"`
	ItemSalePrice      *int64     `gorm:"default: null;" json:"item_sale_price"`
	ItemDescription    *string    `gorm:"type:text;default:null;" json:"item_description"`
	Shippable          bool       `gorm:"not null;default:false;" json:"shippable"`
	Quantity           int        `gorm:"not null; default:0;" json:"quantity"`
	StripeProductApiID string     `gorm:"not null; type:varchar(255)" json:"stripe_product_api_id"`
	StripePriceApiID   string     `gorm:"type:varchar(255);" json:"stripe_price_api_id"`
	CreatedAt          time.Time  `gorm:"not null;" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"not null;" json:"updated_at"`
	DeletedAt          *time.Time `json:"-"`
	db                 *gorm.DB
}

func (i *ShopItem) TableName() string {
	return "shop_items"
}

func NewShopItem(db *gorm.DB, stripeKey string) *ShopItem {
	stripe.Key = stripeKey
	return &ShopItem{db: db}
}

func NewShopItemForUpdate(db *gorm.DB, stripeKey string, stripeProductApiID, stripePriceApiID string) *ShopItem {
	stripe.Key = stripeKey
	return &ShopItem{db: db, StripeProductApiID: stripeProductApiID, StripePriceApiID: stripePriceApiID}
}

func (i *ShopItem) setUpdatedAt() {
	i.UpdatedAt = time.Now()
}

func (i *ShopItem) Create(data *ShopItemCreate) error {
	if data == nil {
		return ErrShopItemCreateBlank
	}

	if err := data.Validate(); err != nil {
		return err
	}

	i.ItemName = data.ItemName
	i.ItemPicture = data.ItemPicture
	i.ItemPrice = data.ItemPrice
	i.ItemSalePrice = data.ItemSalePrice
	i.ItemDescription = data.ItemDescription
	i.Shippable = data.Shippable
	i.Quantity = data.Quantity
	i.CreatedAt = time.Now()
	i.setUpdatedAt()

	stripeProduct, err := data.createStripeProduct(data.ItemName, data.ItemDescription)
	if err != nil {
		log.Printf("error occurred while creating stripe product: %v", err)
		return err
	}

	itemPrice := data.ItemPrice
	if data.ItemSalePrice != nil {
		itemPrice = *data.ItemSalePrice
	}

	productPrice, err := data.createStripeProductPrice(stripeProduct, itemPrice)
	if err != nil {
		log.Printf("error occurred while creating stripe product price: %v", err)
		return err
	}

	insertQuery := `INSERT INTO shop_items (item_name, item_picture, item_price, item_sale_price, item_description, 
		shippable, quantity, stripe_product_api_id, stripe_price_api_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	params := []interface{}{i.ItemName, i.ItemPicture, i.ItemPrice, i.ItemSalePrice, i.ItemDescription, i.Shippable,
		i.Quantity, stripeProduct.ID, productPrice.ID, i.CreatedAt, i.UpdatedAt}

	if err := i.db.Debug().Exec(insertQuery, params...).Error; err != nil {
		log.Printf("error while saving to db: %v\n", err)
		return ErrInternal
	}

	lastID, err := getLastInsertedID(i.db)
	if err != nil {
		return err
	}

	i.ID = lastID
	i.StripeProductApiID = stripeProduct.ID
	i.StripePriceApiID = productPrice.ID
	return nil
}

func (i *ShopItem) Update(data *ShopItemUpdate) error {
	if data == nil {
		return ErrShopItemUpdateBlank
	}

	if err := data.Validate(); err != nil {
		return err
	}

	i.ItemName = data.ItemName
	i.ItemPicture = data.ItemPicture
	i.ItemPrice = data.ItemPrice
	i.ItemSalePrice = data.ItemSalePrice
	i.ItemDescription = data.ItemDescription
	i.Shippable = data.Shippable
	i.Quantity = data.Quantity
	i.setUpdatedAt()

	if _, err := data.updateStripeProduct(i.StripeProductApiID, i.ItemName, i.ItemDescription); err != nil {
		log.Printf("error occurred while updating stripe product: %v", err)
		return err
	}

	itemPrice := i.ItemPrice
	if i.ItemSalePrice != nil {
		itemPrice = *data.ItemSalePrice
	}

	if _, err := data.updateStripeProductPrice(i.StripeProductApiID, itemPrice); err != nil {
		log.Printf("error occurred while updating stripe product price: %v", err)
		return err
	}

	updateQuery := `UPDATE shop_items SET item_name = ?, item_picture = ?, item_price = ?, item_sale_price = ?, item_description = ?, 
		shippable = ?, quantity = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`

	params := []interface{}{i.ItemName, i.ItemPicture, i.ItemPrice, i.ItemSalePrice, i.ItemDescription, i.Shippable,
		i.Quantity, i.UpdatedAt, i.ID}

	if err := i.db.Debug().Exec(updateQuery, params...).Error; err != nil {
		log.Printf("error while updating shop item: %v\n", err)
		return ErrInternal
	}

	return nil
}

func (i *ShopItem) Delete() error {
	currentTime := time.Now()
	i.setUpdatedAt()
	i.DeletedAt = &currentTime

	return nil
}
