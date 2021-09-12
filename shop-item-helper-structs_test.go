package mop_shop

import (
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v72"
	"testing"
)

type StripeProductTest struct {
}

func TestShopItemCreate_Validate(t *testing.T) {
	type fields struct {
		ItemName        string
		ItemPicture     *string
		ItemPrice       int64
		ItemSalePrice   *int64
		ItemDescription *string
		Shippable       bool
		Quantity        int
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name:    "Item name is required",
			fields:  fields{},
			wantErr: ErrItemNameBlank,
		},
		{
			name: "Item price is negative",
			fields: fields{
				ItemName:  "Test item",
				ItemPrice: -20,
			},
			wantErr: ErrShopItemPriceNegative,
		},
		{
			name: "Item sale price is negative",
			fields: fields{
				ItemName:      "Test item",
				ItemPrice:     20,
				ItemSalePrice: stripe.Int64(-209),
			},
			wantErr: ErrShopItemSalePriceNegative,
		},
		{
			name: "Item sale price is negative",
			fields: fields{
				ItemName:      "Test item",
				ItemPrice:     20,
				ItemSalePrice: stripe.Int64(-10),
			},
			wantErr: ErrShopItemSalePriceNegative,
		},
		{
			name: "Item sale price can't be higher than item price",
			fields: fields{
				ItemName:      "Test item",
				ItemPrice:     20,
				ItemSalePrice: stripe.Int64(30),
			},
			wantErr: ErrShopItemSalePriceGreaterThanItemPrice,
		},
		{
			name: "Quantity can't be zero",
			fields: fields{
				ItemName:      "Test item",
				ItemPrice:     20,
				ItemSalePrice: stripe.Int64(10),
				Quantity:      0,
			},
			wantErr: ErrShopItemQuantityZeroOrNegative,
		},
		{
			name: "Quantity can't be zero",
			fields: fields{
				ItemName:      "Test item",
				ItemPrice:     20,
				ItemSalePrice: stripe.Int64(10),
				Quantity:      -20,
			},
			wantErr: ErrShopItemQuantityZeroOrNegative,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ShopItemCreate{
				ItemName:        tt.fields.ItemName,
				ItemPicture:     tt.fields.ItemPicture,
				ItemPrice:       tt.fields.ItemPrice,
				ItemSalePrice:   tt.fields.ItemSalePrice,
				ItemDescription: tt.fields.ItemDescription,
				Shippable:       tt.fields.Shippable,
				Quantity:        tt.fields.Quantity,
			}

			validationErr := c.Validate()
			assert.Equal(t, tt.wantErr, validationErr, "Validate() error = %v, wantErr %v", validationErr, tt.wantErr)
		})
	}
}

func TestShopItemUpdate_Validate(t *testing.T) {
	type fields struct {
		ItemName        string
		ItemPicture     *string
		ItemPrice       int64
		ItemSalePrice   *int64
		ItemDescription *string
		Shippable       bool
		Quantity        int
		stripeProductID string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "All is good",
			fields: fields{
				ItemName:        "OK",
				ItemPicture:     nil,
				ItemPrice:       21,
				ItemSalePrice:   stripe.Int64(20),
				ItemDescription: nil,
				Shippable:       false,
				Quantity:        1,
				stripeProductID: "prod_klwytt545",
			},
			wantErr: nil,
		},
		{
			name: "ShopItemUpdate is not initialized via constructor",
			fields: fields{
				ItemName:        "OK",
				ItemPicture:     nil,
				ItemPrice:       21,
				ItemSalePrice:   stripe.Int64(20),
				ItemDescription: nil,
				Shippable:       false,
				Quantity:        1,
				stripeProductID: "",
			},
			wantErr: ErrShopItemNotInitializedProperly,
		},
		{
			name: "Item name cannot be blank",
			fields: fields{
				ItemName:        "",
				ItemPicture:     nil,
				ItemPrice:       21,
				ItemSalePrice:   stripe.Int64(20),
				ItemDescription: nil,
				Shippable:       false,
				Quantity:        1,
				stripeProductID: "prod_asbn441",
			},
			wantErr: ErrItemNameBlank,
		},
		{
			name: "Item price cannot be negative",
			fields: fields{
				ItemName:        "Test product",
				ItemPicture:     nil,
				ItemPrice:       -21,
				ItemSalePrice:   stripe.Int64(20),
				ItemDescription: nil,
				Shippable:       false,
				Quantity:        1,
				stripeProductID: "prod_asbn441",
			},
			wantErr: ErrShopItemPriceNegative,
		},
		{
			name: "Item sale price cannot be negative",
			fields: fields{
				ItemName:        "Test product",
				ItemPicture:     nil,
				ItemPrice:       21,
				ItemSalePrice:   stripe.Int64(-20),
				ItemDescription: nil,
				Shippable:       false,
				Quantity:        1,
				stripeProductID: "prod_asbn441",
			},
			wantErr: ErrShopItemSalePriceNegative,
		},
		{
			name: "Item sale price cannot be greater than item price",
			fields: fields{
				ItemName:        "Test product",
				ItemPicture:     nil,
				ItemPrice:       21,
				ItemSalePrice:   stripe.Int64(100),
				ItemDescription: nil,
				Shippable:       false,
				Quantity:        1,
				stripeProductID: "prod_asbn441",
			},
			wantErr: ErrShopItemSalePriceGreaterThanItemPrice,
		},
		{
			name: "Item sale price cannot be greater than item price",
			fields: fields{
				ItemName:        "Test product",
				ItemPicture:     nil,
				ItemPrice:       21,
				ItemSalePrice:   nil,
				ItemDescription: nil,
				Shippable:       false,
				Quantity:        -1,
				stripeProductID: "prod_asbn441",
			},
			wantErr: ErrShopItemQuantityZeroOrNegative,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &ShopItemUpdate{
				ItemName:        tt.fields.ItemName,
				ItemPicture:     tt.fields.ItemPicture,
				ItemPrice:       tt.fields.ItemPrice,
				ItemSalePrice:   tt.fields.ItemSalePrice,
				ItemDescription: tt.fields.ItemDescription,
				Shippable:       tt.fields.Shippable,
				Quantity:        tt.fields.Quantity,
				stripeProductID: tt.fields.stripeProductID,
			}
			validationErr := u.Validate()
			assert.Equal(t, tt.wantErr, validationErr, "Validate() error = %v, wantErr %v", validationErr, tt.wantErr)
		})
	}
}
