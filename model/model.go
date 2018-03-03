package model

import (
	"encoding/json"
	"time"

)

type Model struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt,omitempty"`
}

type Slugged struct {
	Slug          string `sql:"index" json:"slug"`
}

func (m *Model) ToJSON() string {
	r, _ := json.Marshal(m)
	return string(r)
}
