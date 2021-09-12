package mop_shop

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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
