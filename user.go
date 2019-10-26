package lrs

import (
	"crypto/rand"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/gosimple/slug"
	"github.com/jbenet/go-base58"
	"github.com/zoonman/gravatar"
	"golang.org/x/crypto/bcrypt"
)

// User represents user entity
type User struct {
	Model
	Email         string          `validator:"email" gorm:"unique_index" json:"email,omitempty"`
	EmailVerified bool            `gorm:"default:false" json:"email_verified,omitempty"`
	Password      string          `json:"-"`
	Hash          string          `json:"-"`
	Name          string          `json:"name"`
	Slug          string          `json:"slug" gorm:"unique_index"`
	Picture       string          `json:"picture"`
	About         string          `json:"about,omitempty"`
	Gender        string          `json:"gender,omitempty"`
	Rank          float32         `json:"rank"`
	Online        bool            `json:"online"`
	Roles         []Role          `json:"roles,omitempty"`
	Profiles      []SocialProfile `json:"profiles,omitempty" gorm:"[]"`
	Devices       []Device        `json:"devices,omitempty"`
	Settings      *Settings       `json:"settings,omitempty"`

	// Status?

	// Available
	// Busy
	// Offline
}

// UserList for list of the users
type UserList []User

// Device describes the device used by the user
type Device struct {
	UserID       uint
	DeviceID     string
	Type         string // browser, phone
	UserAgent    string
	LastIP       net.Addr
	AccessAt     time.Time
	Subscribed   bool
	PushEndpoint string
	PushKeyP256  string
	PushAuth     string
}

// Settings keep user settings under control
type Settings struct {
	// Offline Notifications:
	// 0. No
	// 1. Push only
	// 2. Immediate to email
	// 3. Daily email digests
	// 4. Weekly email digests
	UserID        uint
	Notifications uint `json:"notifications"`
	Timezone      time.Location
}

// SocialProfile represents user profile in social networks
type SocialProfile struct {
	Model
	NetworkID string    `json:"networkId"`
	Network   string    `json:"name"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"-"`
	UserID    uint
}

// Role of the user
type Role struct {
	Role string `json:"role"`
}

// UserAuthData packet
type UserAuthData struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

// returns a string for a given length
func randomString(l uint8) string {
	b := make([]byte, l)
	rand.Read(b)
	return base58.Encode(b)[:l]
}

// SetPassword for the User
func (u *User) SetPassword(password string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		u.Password = string(hashedPassword)
		u.Hash = randomString(16)
	} else {
		logrus.Error(err)
	}
}

// MakeNameFromEmail method
func (u *User) MakeNameFromEmail() string {
	emails := strings.Split(u.Email, "@")
	name := emails[0]

	for _, v := range []string{".", "_", "-"} {
		name = strings.Replace(name, v, " ", -1)
	}
	return strings.Title(name)
}

// MakeSlug method
func (u *User) MakeSlug() {
	slug.MaxLength = 12
	u.Slug = strings.Replace(
		slug.Make(u.Name),
		"-",
		"",
		-1,
	)
}

// MakeGravatarPicture method
func (u *User) MakeGravatarPicture() string {
	return gravatar.Avatar(u.Email, 100)
}

// SafePluck returns sanitized version of User object
func (u *User) SafePluck() User {
	var ru User
	ru.Name = u.Name
	ru.Picture = u.Picture
	ru.Slug = u.Slug
	ru.Rank = u.Rank
	ru.Online = u.Online
	ru.Gender = u.Gender
	ru.ID = u.ID
	ru.CreatedAt = u.CreatedAt
	ru.UpdatedAt = u.UpdatedAt
	return ru
}

// Map applies given function to list of users and
func (ul UserList) Map(f func(User) User) UserList {
	nl := make(UserList, len(ul))
	for i, v := range ul {
		nl[i] = f(v)
	}
	return nl
}
