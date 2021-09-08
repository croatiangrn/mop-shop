package mop_shop

import (
	"gorm.io/gorm"
	"log"
)

func getLastInsertedID(db *gorm.DB) (int, error) {
	lastInsertedID := 0
	lastInsertIDQuery := `SELECT LAST_INSERT_ID()`

	if err := db.Debug().Raw(lastInsertIDQuery).Scan(&lastInsertedID).Error; err != nil {
		log.Printf("error occurred while fetching last insert ID: %v\n", err)
		return 0, ErrInternal
	}

	return lastInsertedID, nil
}
