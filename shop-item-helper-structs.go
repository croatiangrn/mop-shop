package mop_shop

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"gorm.io/gorm"
)

type ShopItemCreate struct {
	ItemName        string  `json:"item_name"`
	ItemPicture     *string `json:"item_picture"`
	ItemPrice       int64   `json:"item_price"`
	ItemSalePrice   *int64  `json:"item_sale_price"`
	ItemDescription *string `json:"item_description"`
	Shippable       bool    `json:"shippable"`
	Quantity        int     `json:"quantity"`
}

func NewShopItemCreate() *ShopItemCreate {
	return &ShopItemCreate{}
}

func (c *ShopItemCreate) createStripeProduct(name string, description *string) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: description,
	}

	return product.New(params)
}

func (c *ShopItemCreate) createStripeProductPrice(product *stripe.Product, unitAmount int64, lookupKey string) (*stripe.Price, error) {
	priceParams := &stripe.PriceParams{
		Product:    stripe.String(product.ID),
		Currency:   stripe.String(string(stripe.CurrencyEUR)),
		UnitAmount: stripe.Int64(unitAmount),
		LookupKey:  stripe.String(lookupKey),
	}

	return price.New(priceParams)
}

func (c *ShopItemCreate) Validate() error {
	if len(c.ItemName) == 0 {
		return ErrItemNameBlank
	}

	if c.ItemPrice < 0 {
		return ErrShopItemPriceNegative
	}

	if c.ItemSalePrice != nil && *c.ItemSalePrice < 0 {
		return ErrShopItemSalePriceNegative
	}

	if c.Quantity < 0 {
		return ErrShopItemQuantityNegative
	}

	return nil
}

type ShopItemUpdate struct {
	ItemName        string  `json:"item_name"`
	ItemPicture     *string `json:"item_picture"`
	ItemPrice       int64   `json:"item_price"`
	ItemSalePrice   *int64  `json:"item_sale_price"`
	ItemDescription *string `json:"item_description"`
	Shippable       bool    `json:"shippable"`
	Quantity        int     `json:"quantity"`
	stripeProductID string
	db              *gorm.DB
}

func NewShopItemUpdate(db *gorm.DB, stripeProductID string) *ShopItemUpdate {
	return &ShopItemUpdate{db: db, stripeProductID: stripeProductID}
}

func (u *ShopItemUpdate) Validate() error {
	if len(u.ItemName) == 0 {
		return ErrItemNameBlank
	}

	if u.ItemPrice < 0 {
		return ErrShopItemPriceNegative
	}

	if u.ItemSalePrice != nil && *u.ItemSalePrice < 0 {
		return ErrShopItemSalePriceNegative
	}

	if u.Quantity < 0 {
		return ErrShopItemQuantityNegative
	}

	return nil
}

func (u *ShopItemUpdate) updateStripeProduct(stripeProductApiID, name string, description *string) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: description,
	}

	return product.Update(stripeProductApiID, params)
}

func (u *ShopItemUpdate) updateStripeProductPrice(stripeProductApiID string, unitAmount int64) (*stripe.Price, error) {
	params := &stripe.PriceParams{
		Product:    stripe.String(stripeProductApiID),
		Currency:   stripe.String(string(stripe.CurrencyEUR)),
		UnitAmount: stripe.Int64(unitAmount),
	}

	// This will create new price instead of updating unit amount, if we want to update unit amount then it has to be
	// done using session authentication (Stripe API)
	return price.New(params)
}
