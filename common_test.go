package mop_shop

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

func Test_getLastInsertedID(t *testing.T) {
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

	type args struct {
		db *gorm.DB
	}

	type expectedMock struct {
		expectedDBError error
	}

	tests := []struct {
		name         string
		args         args
		expectedMock expectedMock
		want         int
		wantErr      error
	}{
		{
			name: "All is good",
			args: args{
				db: database,
			},
			expectedMock: expectedMock{
				expectedDBError: nil,
			},
			want:    0,
			wantErr: nil,
		},
		{
			name: "Internal error handle",
			args: args{
				db: database,
			},
			expectedMock: expectedMock{
				expectedDBError: ErrInternal,
			},
			want:    0,
			wantErr: ErrInternal,
		},
	}

	query := `SELECT LAST_INSERT_ID()`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getRows := sqlmock.NewRows([]string{"LAST_INSERT_ID()"}).AddRow(tt.want)

			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(getRows).WillReturnError(tt.expectedMock.expectedDBError)

			got, validationErr := getLastInsertedID(tt.args.db)
			assert.Equal(t, tt.wantErr, validationErr, "Validate() error = %v, wantErr %v", validationErr, tt.wantErr)

			if got != tt.want {
				t.Errorf("getLastInsertedID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaginationParams_normalize(t *testing.T) {
	type fields struct {
		PerPage int
		Before  int
		After   int
	}
	tests := []struct {
		name           string
		fields         fields
		expectedResult *PaginationParams
	}{
		{
			name: "PerPage negative to positive",
			fields: fields{
				PerPage: -2,
			},
			expectedResult: &PaginationParams{
				PerPage: ItemsPerPageDefault,
			},
		},
		{
			name: "Revert PerPage to max allowed value",
			fields: fields{
				PerPage: ItemsPerPageMax + 1,
			},
			expectedResult: &PaginationParams{
				PerPage: ItemsPerPageMax,
			},
		},
		{
			name: "Revert Before to zero",
			fields: fields{
				PerPage: ItemsPerPageDefault,
				Before:  -234,
			},
			expectedResult: &PaginationParams{
				PerPage: ItemsPerPageDefault,
				Before:  0,
			},
		},
		{
			name: "Revert After to zero",
			fields: fields{
				PerPage: ItemsPerPageDefault,
				After:   -23,
			},
			expectedResult: &PaginationParams{
				PerPage: ItemsPerPageDefault,
				After:   0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PaginationParams{
				PerPage: tt.fields.PerPage,
				Before:  tt.fields.Before,
				After:   tt.fields.After,
			}

			p.normalize()

			areEqual := cmp.Equal(p, tt.expectedResult)
			assert.Equal(t, true, areEqual, "got = %v, expected = %v", p, *tt.expectedResult)
		})
	}
}
