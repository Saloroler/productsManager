package models

import "time"

type Product struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     int       `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductEvent struct {
	Type      string `json:"type"`
	ProductID int64  `json:"product_id"`
	Timestamp string `json:"ts"`
}
