package mop_shop

import "errors"

var (
	ErrItemNameBlank             = errors.New("item_name_cannot_be_blank")
	ErrShopItemQuantityNegative  = errors.New("quantity_cannot_be_negative")
	ErrShopItemPriceNegative     = errors.New("price_cannot_be_negative")
	ErrShopItemSalePriceNegative = errors.New("sale_price_cannot_be_negative")
	ErrInternal                  = errors.New("internal_error")
	ErrShopItemCreateBlank       = errors.New("shop_item_create_data_cannot_be_blank")
	ErrShopItemUpdateBlank       = errors.New("shop_item_update_data_cannot_be_blank")
	ErrInvalidUserID             = errors.New("user_id_cannot_be_zero")
	ErrOrderItemsEmpty           = errors.New("order_items_cannot_be_empty")
	ErrInvalidItemID             = errors.New("item_id_cannot_be_less_or_equal_than_zero")
	ErrInvalidItemQuantity       = errors.New("item_quantity_cannot_be_less_or_equal_than_zero")
)
