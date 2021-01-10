package core

import "time"

type Event struct {
	SRC    string    `json:"src"`
	DataID int       `json:"data-id"`
	Date   time.Time `json:"date"`

	Description string   `json:"description"`
	Images      []string `json:"images"`
}
