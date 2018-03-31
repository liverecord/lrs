package server

type Category struct {
	Model
	Name        string `json:"name"`
	Slug        string `sql:"index" json:"slug"`
	Description string `json:"description" sql:"type:text"`
	Order       int    `json:"order"`
	Active      bool   `json:"active"`
	Updates     int    `json:"updates"`
}
