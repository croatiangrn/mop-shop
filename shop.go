package mop_shop

import "gorm.io/gorm"

type ShopItem struct {
	ID            int      `gorm:"primaryKey" json:"id"`
	ItemName      string   `gorm:"not null;type:varchar(255);" json:"item_name"`
	ItemPicture   *string  `gorm:"default:null;type:varchar(255);" json:"item_picture"`
	ItemPrice     float32  `gorm:"not null;" json:"item_price"`
	ItemSalePrice *float32 `gorm:"default: null;" json:"item_sale_price"`
	db            *gorm.DB
}

func (i *ShopItem) TableName() string {
	return "shop_items"
}
