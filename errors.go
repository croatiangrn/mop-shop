package mop_shop

import "errors"

var (
	ErrItemNameBlank             = errors.New("item_name_cannot_be_blank")
	ErrShopItemQuantityNegative  = errors.New("quantity_cannot_be_negative")
	ErrShopItemPriceNegative     = errors.New("price_cannot_be_negative")
	ErrShopItemSalePriceNegative = errors.New("sale_price_cannot_be_negative")
	ErrInternal                  = errors.New("internal_error")
	ErrShopItemCreateBlank       = errors.New("shop_item_create_data_cannot_be_blank")
)
