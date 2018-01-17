package model

import "time"

type Model struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt,omitempty"`
}
/*
type User struct {
	Model
	Name string
}

type Folder struct {
	Model
	Name string
}

type File struct {
	Model
	Name string
	CategoryID 	  uint
	Category      Category
	Acl           []User `json:"acl" gorm:"many2many:file_acl"`
}
*/
