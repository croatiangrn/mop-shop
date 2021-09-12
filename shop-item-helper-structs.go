package mop_shop

import (
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
)

type ShopItemCreate struct {
	ItemName        string  `json:"item_name"`
	ItemPicture     *string `json:"item_picture"`
	ItemPrice       int64   `json:"item_price"`
	ItemSalePrice   *int64  `json:"item_sale_price"`
	ItemDescription *string `json:"item_description"`
	Shippable       bool    `json:"shippable"`
	Quantity        int     `json:"quantity"`
	uuid            string
}

func NewShopItemCreate() *ShopItemCreate {
	return &ShopItemCreate{}
}

func (c *ShopItemCreate) SetUUID(uuid string) {
	c.uuid = uuid
}

func (c *ShopItemCreate) GetUUID() string {
	if len(c.uuid) == 0 {
		c.SetUUID(uuid.New().String())
	}

	return c.uuid
}

func (c *ShopItemCreate) GetItemName() string {
	return c.ItemName
}

func (c *ShopItemCreate) GetItemPicture() *string {
	return c.ItemPicture
}

func (c *ShopItemCreate) GetItemPrice() int64 {
	return c.ItemPrice
}

func (c *ShopItemCreate) GetItemSalePrice() *int64 {
	return c.ItemSalePrice
}

func (c *ShopItemCreate) GetItemDescription() *string {
	return c.ItemDescription
}

func (c *ShopItemCreate) GetShippable() bool {
	return c.Shippable
}

func (c *ShopItemCreate) GetQuantity() int {
	return c.Quantity
}

func (c *ShopItemCreate) createStripeProduct(name string, description *string) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: description,
		Active:      stripe.Bool(true),
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

	if c.ItemSalePrice != nil {
		if *c.ItemSalePrice < 0 {
			return ErrShopItemSalePriceNegative
		}

		if *c.ItemSalePrice > c.ItemPrice {
			return ErrShopItemSalePriceGreaterThanItemPrice
		}
	}

	if c.Quantity <= 0 {
		return ErrShopItemQuantityZeroOrNegative
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
}

func NewShopItemUpdate(stripeProductID string) *ShopItemUpdate {
	return &ShopItemUpdate{stripeProductID: stripeProductID}
}

func (u *ShopItemUpdate) Validate() error {
	if len(u.stripeProductID) == 0 {
		return ErrShopItemNotInitializedProperly
	}

	if len(u.ItemName) == 0 {
		return ErrItemNameBlank
	}

	if u.ItemPrice < 0 {
		return ErrShopItemPriceNegative
	}

	if u.ItemSalePrice != nil {
		if *u.ItemSalePrice < 0 {
			return ErrShopItemSalePriceNegative
		}

		if *u.ItemSalePrice > u.ItemPrice {
			return ErrShopItemSalePriceGreaterThanItemPrice
		}
	}

	if u.Quantity <= 0 {
		return ErrShopItemQuantityZeroOrNegative
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

func (u *ShopItemUpdate) updateStripeProductPrice(productApiID, priceLookupKey string, unitAmount int64) (*stripe.Price, error) {
	params := &stripe.PriceParams{
		Product:           stripe.String(productApiID),
		Currency:          stripe.String(string(stripe.CurrencyEUR)),
		UnitAmount:        stripe.Int64(unitAmount),
		LookupKey:         stripe.String(priceLookupKey),
		TransferLookupKey: stripe.Bool(true),
	}

	// This will create new price instead of updating unit amount, if we want to update unit amount then it has to be
	// done using session authentication (Stripe API)
	return price.New(params)
}
