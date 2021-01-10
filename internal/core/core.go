package core

import "time"

type Event struct {
	SRC    string    `json:"src"`
	DataID int       `json:"data-id"`
	Date   time.Time `json:"date,omitempty"`

	Description string   `json:"description,omitempty"`
	Images      []string `json:"images,omitempty"`
}
