package lrs

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/liverecord/lrs/common"
	"time"
)

// Topic defines the main forum topic structure
type Topic struct {
	Model
	Slugged
	CategoryID    uint64    `json:"categoryId"`
	Category      Category  `json:"category,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID        uint64    `json:"userId"`
	User          User      `json:"user,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	Title         string    `json:"title"`
	Body          string    `json:"body,omitempty" sql:"type:longtext"`
	Order         int       `json:"order"`
	ACL           []User    `json:"acl" gorm:"many2many:topic_acl;association_autoupdate:false;association_autocreate:false"`
	TotalViews    uint      `json:"total_views,omitempty"`
	TotalComments uint      `json:"total_comments,omitempty"`
	UnreadComments uint     `json:"unread_comments,omitempty"`
	CommentedAt   time.Time `json:"commentedAt,omitempty"`
	Rank          uint      `json:"rank,omitempty"`
	Private       bool      `json:"private"`
	Pinned        bool      `json:"pinned"`
	Spam          bool      `json:"spam"`
	Moderated     bool      `json:"moderated"`
}

// TopicStatus keeps track of topic reads, votes, favorites
type TopicStatus struct {
	TopicID    uint64     `json:"topicId" gorm:"primary_key;AUTO_INCREMENT:false"`
	UserID     uint64     `json:"userId" gorm:"primary_key;AUTO_INCREMENT:false"`
	ReadAt     *time.Time `json:"readAt"`
	NotifiedAt *time.Time `json:"notifiedAt"`
	Vote       int        `json:"vote"`
	Favorite   bool       `json:"favorite"`
	Block      bool       `json:"block"`
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
	t.Title = StripTags(t.Title)
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
