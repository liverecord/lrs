package server

import (
	"crypto/rand"
	"net"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gosimple/slug"
	"github.com/jbenet/go-base58"
	"github.com/zoonman/gravatar"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Model
	Email    string          `validator:"email" gorm:"unique_index" json:"email"`
	Password string          `json:"-"`
	Hash     string          `json:"-"`
	Name     string          `json:"name"`
	Slug     string          `json:"slug" gorm:"unique_index"`
	Picture  string          `json:"picture"`
	About    string          `json:"about,omitempty"`
	Gender   string          `json:"gender,omitempty"`
	Rank     float32         `json:"rank"`
	Online   bool            `json:"online"`
	Roles    []Role          `json:"roles,omitempty"`
	Profiles []SocialProfile `json:"profiles,omitempty" gorm:"[]"`
	Devices  []Devices
	Settings Settings

	// Status?

	// Available
	// Busy
	// Offline
}

type Devices struct {
	UserID       uint
	DeviceID     string
	Type         string // browser, phone
	UserAgent    string
	LastIp       net.Addr
	AccessAt     time.Time
	Subscribed   bool
	PushEndpoint string
	PushKeyP256  string
	PushAuth     string
}

type Settings struct {
	// Offline Notifications:
	// 0. No
	// 1. Push only
	// 2. Immediate to email
	// 3. Daily email digests
	// 4. Weekly email digests
	UserID        uint
	Notifications uint
	Timezone      time.Location
}

type SocialProfile struct {
	Model
	NetworkId string    `json:"network_id"`
	Network   string    `json:"name"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"-"`
	UserId    uint
}

type Role struct {
	Role string `json:"role"`
}

type UserAuthData struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

func randomString(l uint8) string {
	b := make([]byte, l)
	rand.Read(b)
	return base58.Encode(b)[:l]
}

func (u *User) SetPassword(password string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		u.Password = string(hashedPassword)
		u.Hash = randomString(16)
	} else {
		logrus.Error(err)
	}
}

func (u *User) MakeNameFromEmail() string {
	emails := strings.Split(u.Email, "@")
	name := emails[0]

	for _, v := range []string{".", "_", "-"} {
		name = strings.Replace(name, v, " ", -1)
	}
	return strings.Title(name)
}

func (u *User) MakeSlug() {
	slug.MaxLength = 12
	u.Slug = strings.Replace(
		slug.Make(u.Name),
		"-",
		"",
		-1,
	)
}

func (u *User) MakeGravatarPicture() string {
	return gravatar.Avatar(u.Email, 100)
}
