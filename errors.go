package mop_shop

import "errors"

var (
	ErrItemNameBlank                            = errors.New("item_name_cannot_be_blank")
	ErrShopItemQuantityZeroOrNegative           = errors.New("quantity_cannot_be_zero_or_negative")
	ErrShopItemPriceNegative                    = errors.New("price_cannot_be_negative")
	ErrShopItemSalePriceNegative                = errors.New("sale_price_cannot_be_negative")
	ErrShopItemSalePriceGreaterThanItemPrice    = errors.New("sale_price_cannot_be_greater_than_price")
	ErrInternal                                 = errors.New("internal_error")
	ErrShopItemCreateBlank                      = errors.New("shop_item_create_data_cannot_be_blank")
	ErrShopItemUpdateBlank                      = errors.New("shop_item_update_data_cannot_be_blank")
	ErrInvalidUserID                            = errors.New("user_id_cannot_be_zero")
	ErrOrderItemsEmpty                          = errors.New("order_items_cannot_be_empty")
	ErrInvalidItemID                            = errors.New("item_id_cannot_be_less_or_equal_than_zero")
	ErrInvalidItemQuantity                      = errors.New("item_quantity_cannot_be_less_or_equal_than_zero")
	ErrOrderDataBlank                           = errors.New("order_data_cannot_be_blank")
	ErrSomeItemsDoNotExist                      = errors.New("some_items_do_not_exist")
	ErrInvalidUserOrderID                       = errors.New("user_order_id_cannot_be_zero")
	ErrCommittingTransaction                    = errors.New("could_not_commit_db_transaction")
	ErrParsingURL                               = errors.New("could_not_parse_current_url")
	ErrInsufficientProductStockAmount           = errors.New("insufficient_product_stock_amount")
	ErrShopItemValidationNotInitializedProperly = errors.New("shop_item_validation_is_not_initialized_via_constructor")
)
