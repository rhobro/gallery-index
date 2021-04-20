package core

import "time"

type Event struct {
	SRC    string    `json:"src"`
	Date   time.Time `json:"date,omitempty"`

	Description string   `json:"description,omitempty"`
	Images      []string `json:"images,omitempty"`
}
