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

	if _, err := data.createStripeProductPrice(stripeProduct, itemPrice); err != nil {
		log.Printf("error occurred while creating stripe product price: %v", err)
		return err
	}

	insertQuery := `INSERT INTO shop_items (item_name, item_picture, item_price, item_sale_price, item_description, 
		shippable, quantity, stripe_product_api_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	params := []interface{}{data.ItemName, data.ItemPicture, data.ItemPrice, data.ItemSalePrice, data.ItemDescription, data.Shippable,
		data.Quantity, stripeProduct.ID, i.CreatedAt, i.UpdatedAt}

	if err := i.db.Debug().Exec(insertQuery, params...).Error; err != nil {
		log.Printf("error while saving to db: %v\n", err)
		return ErrInternal
	}

	lastID, err := getLastInsertedID(i.db)
	if err != nil {
		return err
	}

	i.ID = lastID
	return nil
}

func (i *ShopItem) Update() error {
	i.setUpdatedAt()

	return nil
}

func (i *ShopItem) Delete() error {
	currentTime := time.Now()
	i.setUpdatedAt()
	i.DeletedAt = &currentTime

	return nil
}
