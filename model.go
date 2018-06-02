package lrs

import (
	"encoding/json"
	"time"
)

// Model is a prototype structure
type Model struct {
	ID        uint64     `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt,omitempty"`
	UpdatedAt time.Time  `json:"updatedAt,omitempty"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt,omitempty"`
}

// Slugged is prototype for all slugged objects
type Slugged struct {
	Slug string `sql:"index" json:"slug"`
}

// ToJSON returns a stringified JSON for a given Model
func (m *Model) ToJSON() string {
	r, _ := json.Marshal(m)
	return string(r)
}
