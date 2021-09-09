package mop_shop

import "time"

type CreateUserOrder struct {
	userID     int
	Items      []CreateUserOrderItem `json:"items"`
	totalPrice float32
	createdAt  time.Time
}

func NewCreateUserOrder(userID int) *CreateUserOrder {
	return &CreateUserOrder{userID: userID}
}

func (c *CreateUserOrder) validate() error {
	if c.userID == 0 {
		return ErrInvalidUserID
	}

	if len(c.Items) == 0 {
		return ErrOrderItemsEmpty
	}

	for i := range c.Items {
		if c.Items[i].ItemID <= 0 {
			return ErrInvalidItemID
		}

		if c.Items[i].Quantity <= 0 {
			return ErrInvalidItemQuantity
		}
	}

	return nil
}

type CreateUserOrderItem struct {
	ItemID    int `json:"item_id"`
	Quantity  int `json:"quantity"`
	itemPrice float32
}
