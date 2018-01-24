package model

import (
	"strings"

	"github.com/gosimple/slug"
	"github.com/liverecord/server/common"
)

type Topic struct {
	Model
	CategoryID    uint
	Category      Category
	Title         string `json:"title"`
	Slug          string `sql:"index" json:"slug"`
	Body          string `json:"body"`
	Order         int    `json:"order"`
	Acl           []User `json:"acl" gorm:"many2many:topic_acl"`
	TotalViews    uint32 `json:"total_views"`
	TotalComments uint32 `json:"total_comments"`
	Rank          uint32 `json:"rank"`
}

func (t *Topic) BeforeCreate() (err error) {
	t.Slug = slug.Make(t.Title)
	return
}

func (t *Topic) BeforeSave() (err error) {
	t.Title = strings.TrimSpace(t.Title)
	t.Title = strings.Title(t.Title)
	t.Body = common.FilterHtml(t.Body)
	return
}
