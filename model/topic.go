package model

type Topic struct {
	Model
	Category      Category
	Title         string `json:"title"`
	Slug          string `sql:"index" json:"slug"`
	Body          string `json:"description"`
	Order         int    `json:"order"`
	User          User   `json:"user_id"`
	Acl           []User `json:"acl"`
	TotalViews    uint32 `json:"total_views"`
	TotalComments uint32 `json:"total_comments"`
	Rank          uint32 `json:"rank"`
}
