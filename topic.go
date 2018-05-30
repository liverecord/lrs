package lrs

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/liverecord/lrs/common"
)

// Topic defines the main forum topic structure
type Topic struct {
	Model
	Slugged
	CategoryID    uint     `json:"categoryId"`
	Category      Category `json:"category,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID        uint     `json:"userId"`
	User          User     `json:"user,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	Title         string   `json:"title"`
	Body          string   `json:"body,omitempty" sql:"type:longtext"`
	Order         int      `json:"order"`
	ACL           []User   `json:"acl" gorm:"many2many:topic_acl;association_autoupdate:false;association_autocreate:false"`
	TotalViews    uint32   `json:"total_views,omitempty"`
	TotalComments uint32   `json:"total_comments,omitempty"`
	Rank          uint32   `json:"rank,omitempty"`
	Private       bool     `json:"private"`
	Pinned        bool     `json:"pinned"`
}

func makeUniqueSlug(s *string, db *gorm.DB, i uint) {
	rdb := db.Where("slug = ?", s).First(&Topic{})
	if !rdb.RecordNotFound() {
		i++
		*s = fmt.Sprintf("%s%d", strings.TrimRight(*s, "0123456789"), i)
		makeUniqueSlug(s, db, i)
	}
}

// BeforeCreate hook
func (t *Topic) BeforeCreate(scope *gorm.Scope) (err error) {
	t.Title = strings.TrimSpace(t.Title)
	if len(t.Title) == 0 {
		t.Slug = "unknown"
	} else {
		t.Slug = slug.Make(t.Title)
	}
	makeUniqueSlug(&t.Slug, scope.DB(), 0)
	return
}

// BeforeSave hook
func (t *Topic) BeforeSave() (err error) {
	t.Title = strings.TrimSpace(t.Title)
	t.Title = strings.Title(t.Title)
	t.Title = common.StripTags(t.Title)
	t.Body = common.SanitizeHtml(t.Body)
	return
}

// SafeTopic returns sanitized version of the topic
func (t *Topic) SafeTopic() *Topic {
	if t.ACL == nil {
		t.ACL = make([]User, 0)
	} else {
		for i, u := range t.ACL {
			t.ACL[i] = u.SafePluck()
		}
	}
	t.User = t.User.SafePluck()
	return t
}
