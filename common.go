package mop_shop

import (
	"gorm.io/gorm"
	"log"
)

const ItemsPerPageMax = 50
const ItemsPerPageDefault = 20

func getLastInsertedID(db *gorm.DB) (int, error) {
	lastInsertedID := 0
	lastInsertIDQuery := `SELECT LAST_INSERT_ID()`

	if err := db.Debug().Raw(lastInsertIDQuery).Scan(&lastInsertedID).Error; err != nil {
		log.Printf("error occurred while fetching last insert ID: %v\n", err)
		return 0, ErrInternal
	}

	return lastInsertedID, nil
}

type PaginationParams struct {
	PerPage int
	Before  int
	After   int
}

func (p *PaginationParams) normalize() {
	if p.PerPage <= 0 {
		p.PerPage = ItemsPerPageDefault
	}

	if p.PerPage > ItemsPerPageMax {
		p.PerPage = ItemsPerPageMax
	}

	if p.Before < 0 {
		p.Before = 0
	}

	if p.After < 0 {
		p.After = 0
	}
}

type PaginationResponse struct {
	CursorBefore *string `json:"cursor_before"`
	CursorAfter  *string `json:"cursor_after"`
	Before       *string `json:"before"`
	After        *string `json:"after"`
}
