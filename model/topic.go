package model

import (
	"fmt"
	"strings"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/liverecord/server/common"
)

type Topic struct {
	Model
	Slugged
	CategoryID    uint `json:"categoryId"`
	Category      Category `json:"category,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID		  uint `json:"userId"`
	User      	  User `json:"user,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	Title         string `json:"title"`
	Body          string `json:"body,omitempty" sql:"type:longtext"`
	Order         int    `json:"order"`
	Acl           []User `json:"acl,omitempty" gorm:"many2many:topic_acl;association_autoupdate:false;association_autocreate:false"`
	TotalViews    uint32 `json:"total_views,omitempty"`
	TotalComments uint32 `json:"total_comments,omitempty"`
	Rank          uint32 `json:"rank,omitempty"`
	Private       bool   `json:"private"`
}

func makeUniqueSlug(s *string, db *gorm.DB, i uint)  {
	rdb := db.Where("slug = ?", s).First(&Topic{})
	if !rdb.RecordNotFound() {
		i++
		*s = fmt.Sprintf("%s%d", strings.TrimRight(*s, "0123456789"), i)
		makeUniqueSlug(s, db, i)
	}
}

func (t *Topic) BeforeCreate(scope *gorm.Scope) (err error) {
	t.Slug = slug.Make(t.Title)
	makeUniqueSlug(&t.Slug, scope.DB(), 0)
	return
}

func (t *Topic) BeforeSave() (err error) {
	t.Title = strings.TrimSpace(t.Title)
	t.Title = strings.Title(t.Title)
	t.Title = common.StripTags(t.Title)
	t.Body = common.SanitizeHtml(t.Body)
	return
}
