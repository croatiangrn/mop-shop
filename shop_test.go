package mop_shop

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v72"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
)

func TestShopItem_FindOneByID(t *testing.T) {
	dbTest, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer dbTest.Close()

	dialector := mysql.New(mysql.Config{
		DSN:                       "sqlmock_db_0",
		DriverName:                "postgres",
		Conn:                      dbTest,
		SkipInitializeWithVersion: true,
	})

	database, _ := gorm.Open(dialector, &gorm.Config{})

	type fields struct {
		ID                         int
		ItemName                   string
		ItemPicture                *string
		ItemPrice                  int64
		ItemSalePrice              *int64
		ItemDescription            *string
		Shippable                  bool
		Quantity                   int
		StripeProductApiID         string
		UniqueStripePriceLookupKey string
		CreatedAt                  time.Time
		UpdatedAt                  time.Time
		DeletedAt                  *time.Time
		db                         *gorm.DB
	}

	type args struct {
		shopItemID int
	}

	type expectedMock struct {
		expectedDBError error
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		expectedMock expectedMock
		wantErr      error
	}{
		{
			name: "All is good",
			fields: fields{
				ID:                         0,
				ItemName:                   "",
				ItemPicture:                nil,
				ItemPrice:                  0,
				ItemSalePrice:              nil,
				ItemDescription:            nil,
				Shippable:                  false,
				Quantity:                   0,
				StripeProductApiID:         "",
				UniqueStripePriceLookupKey: "",
				CreatedAt:                  time.Now(),
				UpdatedAt:                  time.Now(),
				DeletedAt:                  nil,
				db:                         database,
			},
			args: args{
				shopItemID: 0,
			},
			expectedMock: expectedMock{
				expectedDBError: nil,
			},
			wantErr: nil,
		},
		{
			name: "Record not found",
			fields: fields{
				ID:                         0,
				ItemName:                   "",
				ItemPicture:                nil,
				ItemPrice:                  0,
				ItemSalePrice:              nil,
				ItemDescription:            nil,
				Shippable:                  false,
				Quantity:                   0,
				StripeProductApiID:         "",
				UniqueStripePriceLookupKey: "",
				CreatedAt:                  time.Now(),
				UpdatedAt:                  time.Now(),
				DeletedAt:                  nil,
				db:                         database,
			},
			args: args{
				shopItemID: 0,
			},
			expectedMock: expectedMock{
				expectedDBError: gorm.ErrRecordNotFound,
			},
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name: "Internal error occurred",
			fields: fields{
				ID:                         0,
				ItemName:                   "",
				ItemPicture:                nil,
				ItemPrice:                  0,
				ItemSalePrice:              nil,
				ItemDescription:            nil,
				Shippable:                  false,
				Quantity:                   0,
				StripeProductApiID:         "",
				UniqueStripePriceLookupKey: "",
				CreatedAt:                  time.Now(),
				UpdatedAt:                  time.Now(),
				DeletedAt:                  nil,
				db:                         database,
			},
			args: args{
				shopItemID: 0,
			},
			expectedMock: expectedMock{
				expectedDBError: gorm.ErrInvalidDB,
			},
			wantErr: ErrInternal,
		},
	}

	getUserQuery := `SELECT * FROM shop_items WHERE id = ? AND deleted_at IS NULL`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &ShopItem{
				ID:                         tt.fields.ID,
				ItemName:                   tt.fields.ItemName,
				ItemPicture:                tt.fields.ItemPicture,
				ItemPrice:                  tt.fields.ItemPrice,
				ItemSalePrice:              tt.fields.ItemSalePrice,
				ItemDescription:            tt.fields.ItemDescription,
				Shippable:                  tt.fields.Shippable,
				Quantity:                   tt.fields.Quantity,
				StripeProductApiID:         tt.fields.StripeProductApiID,
				UniqueStripePriceLookupKey: tt.fields.UniqueStripePriceLookupKey,
				CreatedAt:                  tt.fields.CreatedAt,
				UpdatedAt:                  tt.fields.UpdatedAt,
				DeletedAt:                  tt.fields.DeletedAt,
				db:                         tt.fields.db,
			}

			getRows := sqlmock.NewRows([]string{"id"}).AddRow(tt.fields.ID)

			mock.ExpectQuery(regexp.QuoteMeta(getUserQuery)).WillReturnRows(getRows).WillReturnError(tt.expectedMock.expectedDBError)

			methodErr := i.FindOneByID(tt.args.shopItemID)
			assert.Equal(t, tt.wantErr, methodErr, "FindOneByID() error = %v, wantErr %v", methodErr, tt.wantErr)

		})
	}
}

type ShopItemCreateTest struct {
	ItemName        string  `json:"item_name"`
	ItemPicture     *string `json:"item_picture"`
	ItemPrice       int64   `json:"item_price"`
	ItemSalePrice   *int64  `json:"item_sale_price"`
	ItemDescription *string `json:"item_description"`
	Shippable       bool    `json:"shippable"`
	Quantity        int     `json:"quantity"`
	uuid            string
}

func (s ShopItemCreateTest) GetUUID() string {
	return s.uuid
}

func (s ShopItemCreateTest) GetItemName() string {
	return s.ItemName
}

func (s ShopItemCreateTest) GetItemPicture() *string {
	return s.ItemPicture
}

func (s ShopItemCreateTest) GetItemPrice() int64 {
	return s.ItemPrice
}

func (s ShopItemCreateTest) GetItemSalePrice() *int64 {
	return s.ItemSalePrice
}

func (s ShopItemCreateTest) GetItemDescription() *string {
	return s.ItemDescription
}

func (s ShopItemCreateTest) GetShippable() bool {
	return s.Shippable
}

func (s ShopItemCreateTest) GetQuantity() int {
	return s.Quantity
}

func (s ShopItemCreateTest) createStripeProduct(name string, description *string) (*stripe.Product, error) {
	return &stripe.Product{}, nil
}

func (s ShopItemCreateTest) createStripeProductPrice(product *stripe.Product, unitAmount int64, lookupKey string) (*stripe.Price, error) {
	return &stripe.Price{}, nil
}

func (s ShopItemCreateTest) Validate() error {
	if len(s.ItemName) == 0 {
		return ErrItemNameBlank
	}

	if s.ItemPrice < 0 {
		return ErrShopItemPriceNegative
	}

	if s.ItemSalePrice != nil {
		if *s.ItemSalePrice < 0 {
			return ErrShopItemSalePriceNegative
		}

		if *s.ItemSalePrice > s.ItemPrice {
			return ErrShopItemSalePriceGreaterThanItemPrice
		}
	}

	if s.Quantity <= 0 {
		return ErrShopItemQuantityZeroOrNegative
	}

	return nil
}

func TestShopItem_Create(t *testing.T) {
	dbTest, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer dbTest.Close()

	dialector := mysql.New(mysql.Config{
		DSN:                       "sqlmock_db_0",
		DriverName:                "postgres",
		Conn:                      dbTest,
		SkipInitializeWithVersion: true,
	})

	database, _ := gorm.Open(dialector, &gorm.Config{})

	type fields struct {
		ID                         int
		ItemName                   string
		ItemPicture                *string
		ItemPrice                  int64
		ItemSalePrice              *int64
		ItemDescription            *string
		Shippable                  bool
		Quantity                   int
		StripeProductApiID         string
		UniqueStripePriceLookupKey string
		CreatedAt                  time.Time
		UpdatedAt                  time.Time
		DeletedAt                  *time.Time
		db                         *gorm.DB
	}
	type args struct {
		data shopItemCreateInterface
	}

	type expectedMock struct {
		expectedDBError error
		expectQuery     bool
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		expectedMock expectedMock
		wantErr      error
	}{
		{
			name:   "Shop item not initialized via constructor",
			fields: fields{},
			args: args{
				data: nil,
			},
			expectedMock: expectedMock{
				expectedDBError: nil,
			},
			wantErr: ErrShopItemNotInitializedProperly,
		},
		{
			name: "Param data cannot be nil",
			fields: fields{
				db: database,
			},
			args: args{
				data: nil,
			},
			expectedMock: expectedMock{
				expectedDBError: nil,
			},
			wantErr: ErrShopItemCreateBlank,
		},
		{
			name: "ErrInternal error",
			fields: fields{
				db: database,
			},
			args: args{
				data: &ShopItemCreateTest{
					ItemName:        "Test item",
					ItemPicture:     nil,
					ItemPrice:       0,
					ItemSalePrice:   nil,
					ItemDescription: nil,
					Shippable:       false,
					Quantity:        2,
				},
			},
			expectedMock: expectedMock{
				expectedDBError: ErrInternal,
				expectQuery:     true,
			},
			wantErr: ErrInternal,
		},
	}

	insertQuery := `INSERT INTO shop_items (item_name, item_picture, item_price, item_sale_price, item_description, shippable, quantity, stripe_product_api_id, unique_stripe_price_lookup_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &ShopItem{
				ID:                         tt.fields.ID,
				ItemName:                   tt.fields.ItemName,
				ItemPicture:                tt.fields.ItemPicture,
				ItemPrice:                  tt.fields.ItemPrice,
				ItemSalePrice:              tt.fields.ItemSalePrice,
				ItemDescription:            tt.fields.ItemDescription,
				Shippable:                  tt.fields.Shippable,
				Quantity:                   tt.fields.Quantity,
				StripeProductApiID:         tt.fields.StripeProductApiID,
				UniqueStripePriceLookupKey: tt.fields.UniqueStripePriceLookupKey,
				CreatedAt:                  tt.fields.CreatedAt,
				UpdatedAt:                  tt.fields.UpdatedAt,
				DeletedAt:                  tt.fields.DeletedAt,
				db:                         tt.fields.db,
			}

			currentTime := time.Now()
			if tt.expectedMock.expectQuery {
				mock.ExpectExec(insertQuery).WithArgs(tt.args.data.GetItemName(), i.ItemPicture, i.ItemPrice, i.ItemSalePrice,
					i.ItemDescription, i.Shippable, tt.args.data.GetQuantity(), i.StripeProductApiID, tt.args.data.GetUUID(),
					currentTime, currentTime).WillReturnError(tt.expectedMock.expectedDBError)
			}

			validationErr := i.Create(tt.args.data, currentTime)
			assert.Equal(t, tt.wantErr, validationErr, "Create() error = %v, wantErr %v", validationErr, tt.wantErr)
		})
	}
}
