package model

type Category struct {
	Model
	Name        string `json:"name"`
	Slug        string `sql:"index" json:"slug"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}
