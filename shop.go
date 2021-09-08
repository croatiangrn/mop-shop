package mop_shop

import (
	"gorm.io/gorm"
	"time"
)

type ShopItem struct {
	ID                 int        `gorm:"primaryKey" json:"id"`
	ItemName           string     `gorm:"not null;type:varchar(255);" json:"item_name"`
	ItemPicture        *string    `gorm:"default:null;type:varchar(255);" json:"item_picture"`
	ItemPrice          float32    `gorm:"not null;" json:"item_price"`
	ItemSalePrice      *float32   `gorm:"default: null;" json:"item_sale_price"`
	ItemDescription    *string    `gorm:"type:text;default:null;" json:"item_description"`
	Shippable          bool       `gorm:"not null;default:false;" json:"shippable"`
	Quantity           int        `gorm:"not null; default:0;" json:"quantity"`
	StripeProductApiID *string    `gorm:"type:varchar(255)" json:"stripe_product_api_id"`
	CreatedAt          time.Time  `gorm:"not null;" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"not null;" json:"updated_at"`
	DeletedAt          *time.Time `json:"-"`
	db                 *gorm.DB
}

func (i *ShopItem) TableName() string {
	return "shop_items"
}

func NewShopItem(db *gorm.DB) *ShopItem {
	return &ShopItem{db: db}
}

func (i *ShopItem) Validate() error {
	return nil
}

func (i *ShopItem) Create() error {
	return nil
}

func (i *ShopItem) Update() error {
	return nil
}

func (i *ShopItem) Delete() error {
	return nil
}
