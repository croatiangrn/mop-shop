package mop_shop

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ShopItemForResponse struct {
	ID             int      `json:"id"`
	ItemName       string   `json:"item_name"`
	ItemPicture    *string  `json:"item_picture"`
	ItemPriceInt64 *int64   `json:"item_price_int_64,omitempty"`
	ItemPrice      *float64 `json:"item_price"`
	// ItemCurrency is a virtual field
	ItemCurrency       string   `gorm:"-" json:"item_currency"`
	ItemSalePriceInt64 *int64   `json:"item_sale_price_int_64,omitempty"`
	ItemSalePrice      *float64 `json:"item_sale_price"`
	ItemDescription    *string  `json:"item_description"`
	Shippable          bool     `json:"shippable"`
	Quantity           *int     `json:"quantity"`
}

// GetShopItemsForFrontend returns all shop items from DB, if ``isAuthorized`` is false then
// these fields will always be nil:
//
// - ShopItemForResponse.ItemSalePrice
//
// - ShopItemForResponse.Quantity
//
// Reason for that is to force users to create an account and see full shop item info
func GetShopItemsForFrontend(isAuthorized bool, currency string, paginationParams PaginationParams, req *http.Request, db *gorm.DB) ([]ShopItemForResponse, *PaginationResponse, error) {
	var shopQuery strings.Builder
	var params []interface{}

	shopQuery.WriteString(`SELECT 
			id, item_name, item_picture, item_price AS item_price_int_64,
			item_sale_price AS item_sale_price_int_64, item_description, 
			shippable, quantity
		FROM shop_items 
		WHERE deleted_at IS NULL `)

	switch {
	case paginationParams.Before == 0 && paginationParams.After == 0:
		shopQuery.WriteString(`AND id > ? ORDER BY id DESC LIMIT ?`)
		params = append(params, 0)
	case paginationParams.After > 0:
		shopQuery.WriteString(`AND id <= ? ORDER BY id DESC LIMIT ?`)
		params = append(params, paginationParams.After)
	case paginationParams.Before > 0:
		shopQuery.WriteString(`AND id >= ? ORDER BY id ASC LIMIT ?`)
		params = append(params, paginationParams.Before)
	}

	params = append(params, paginationParams.PerPage+1)

	var data []ShopItemForResponse
	if err := db.Debug().Raw(shopQuery.String(), params...).Scan(&data).Error; err != nil {
		log.Printf("error while getting user orders: %v\n", err)
		return nil, nil, ErrInternal
	}

	if isAuthorized {
		for i := range data {
			data[i].ItemCurrency = currency
			if data[i].ItemPriceInt64 != nil && *data[i].ItemPriceInt64 != 0 {
				price, _ := decimal.New(*data[i].ItemPriceInt64, -2).Float64()
				data[i].ItemPrice = &price
				data[i].ItemPriceInt64 = nil
			}

			if data[i].ItemSalePriceInt64 != nil && *data[i].ItemSalePriceInt64 != 0 {
				salePrice, _ := decimal.New(*data[i].ItemSalePriceInt64, -2).Float64()
				data[i].ItemSalePrice = &salePrice
				data[i].ItemSalePriceInt64 = nil
			}
		}
	} else {
		for i := range data {
			data[i].ItemCurrency = currency
			if data[i].ItemPriceInt64 != nil && *data[i].ItemPriceInt64 != 0 {
				price, _ := decimal.New(*data[i].ItemPriceInt64, -2).Float64()
				data[i].ItemPrice = &price
				data[i].ItemPriceInt64 = nil
			}
			data[i].ItemSalePrice = nil
			data[i].ItemSalePriceInt64 = nil
			data[i].Quantity = nil
		}
	}

	if len(data) == 0 {
		data = []ShopItemForResponse{}
	}

	pages := PaginationResponse{}

	parsedURL, err := url.Parse(getCurrentURL(req))
	if err != nil {
		return nil, nil, ErrParsingURL
	}

	beforeQ := parsedURL.Query()
	afterQ := parsedURL.Query()

	switch {
	case len(data) == paginationParams.PerPage+1:

		if paginationParams.After > 0 {
			beforeId := strconv.Itoa(data[0].ID + 1)

			beforeQ.Del("after")
			beforeQ.Set("before", beforeId)
			parsedURL.RawQuery = beforeQ.Encode()
			cursorBefore := parsedURL.String()

			pages.CursorBefore = &cursorBefore
			pages.Before = &beforeId
		} else if paginationParams.Before > 0 {
			beforeId := strconv.Itoa(data[0].ID)

			beforeQ.Del("after")
			beforeQ.Set("before", beforeId)
			parsedURL.RawQuery = beforeQ.Encode()
			cursorBefore := parsedURL.String()

			pages.Before = &beforeId
			pages.CursorBefore = &cursorBefore
		}

		afterId := strconv.Itoa(data[len(data)-1].ID)

		afterQ.Set("after", afterId)
		parsedURL.RawQuery = afterQ.Encode()
		cursorAfter := parsedURL.String()

		pages.CursorAfter = &cursorAfter
		pages.After = &afterId

		if paginationParams.Before > 0 {
			afterId := strconv.Itoa(data[len(data)-1].ID - 1)

			afterQ.Del("before")
			afterQ.Set("after", afterId)
			parsedURL.RawQuery = afterQ.Encode()
			cursorAfter := parsedURL.String()

			pages.After = &afterId
			pages.CursorAfter = &cursorAfter
			data = data[1:]
		} else {
			data = data[:len(data)-1]
		}
	case len(data) <= paginationParams.PerPage && paginationParams.Before > 0:
		afterId := strconv.Itoa(data[len(data)-1].ID - 1)

		afterQ.Del("before")
		afterQ.Set("after", afterId)
		parsedURL.RawQuery = afterQ.Encode()
		cursorAfter := parsedURL.String()

		pages.After = &afterId
		pages.CursorAfter = &cursorAfter
	case len(data) <= paginationParams.PerPage && paginationParams.After > 0:
		beforeId := strconv.Itoa(data[0].ID + 1)

		beforeQ.Del("after")
		beforeQ.Set("before", beforeId)
		parsedURL.RawQuery = beforeQ.Encode()
		cursorBefore := parsedURL.String()

		pages.Before = &beforeId
		pages.CursorBefore = &cursorBefore
	}

	return data, &pages, nil
}
