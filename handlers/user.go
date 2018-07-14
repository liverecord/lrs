package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	. "github.com/liverecord/lrs"
	"github.com/liverecord/lrs/mailer"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/bcrypt"
)

// UserLoginData for jwt auth
type UserLoginData struct {
	Jwt  string `json:"jwt"`
	User User   `json:"user"`
}

// UserInfoRequest to get user data
type UserInfoRequest struct {
	Slug string `json:"slug"`
}

// UsersSearchRequest request about users
type UsersSearchRequest struct {
	Term         string `json:"term"`
	ExcludeUsers []uint `json:"exclude"`
}

type ErrorResponse struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

const (
	// GeneralError test
	GeneralError = 0
	// WrongPassword
	WrongPassword = 1
	// NoPasswordReset
	NoPasswordReset = 2
)
// Auth used to authorized user
func (Ctx *ConnCtx) Auth(frame Frame) {
	if Ctx.IsAuthorized() {
		Ctx.Logger.Warningf("Authorized authentication, %v", frame)
		return
	}
	var authData UserAuthData
	frame.BindJSON(&authData)
	Ctx.Logger.Debugf("AuthData: %v", authData)
	var user User
	authData.Email = strings.ToLower(authData.Email)
	Ctx.Db.Where("email = ?", authData.Email).First(&user)
	if user.ID > 0 {
		err := bcrypt.CompareHashAndPassword(
			S2BA(user.Password),
			S2BA(authData.Password))
		if err == nil {
			// we are cool, password is correct
			Ctx.User = &user
			Ctx.respondWithToken(user)
		} else {
			Ctx.Pool.Write(Ctx.Ws, NewFrame(AuthErrorFrame, ErrorResponse{Code:WrongPassword, Message: err.Error()}, frame.RequestID))
			Ctx.Logger.WithError(err).Errorf("Cannot authorize user %s %v", user.Email, err)
		}
		return
	}
	// onboard new user
	user.Email = authData.Email
	user.Name = StripTags(user.MakeNameFromEmail())
	user.Roles = []Role{}
	user.MakeSlug()
	user.SetPassword(authData.Password)
	user.Picture = user.MakeGravatarPicture()
	Ctx.Db.Save(&user)
	Ctx.User = &user
	Ctx.respondWithToken(user)
}

func (Ctx *ConnCtx) respondWithToken(user User) {
	uld, err := Ctx.generateToken(user)
	if err == nil {
		Ctx.Pool.Write(Ctx.Ws, NewFrame(AuthFrame, uld, ""))
		return
	}
	Ctx.Logger.WithError(err).Printf("Cannot generate token %v", err)
}

func (Ctx *ConnCtx) generateToken(user User) (UserLoginData, error) {
	// Create the token
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"uid":  user.ID,
			"hash": user.Hash,
			"exp":  time.Now().Add(time.Hour * 24 * 30).Unix(),
		})
	// Sign and get the complete encoded token as a string
	var uld UserLoginData
	var err error
	uld.User = user
	uld.Jwt, err = token.SignedString(Ctx.Cfg.JwtSignature)
	return uld, err
}

// AuthorizeJWT for checking the JWT token
func (Ctx *ConnCtx) AuthorizeJWT(tokenString string) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return Ctx.Cfg.JwtSignature, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["uid"], claims["exp"])
		Ctx.Logger.WithFields(logrus.Fields(claims)).Debug("Debugged tokens")
		var user User
		Ctx.Db.Where("id = ? AND hash = ?", claims["uid"], claims["hash"]).First(&user)
		if user.ID != 0 {
			Ctx.respondWithToken(user)
			Ctx.User = &user
		}
		return
	}
	Ctx.Logger.Error(err)
}

// ResetPassword Resets user password
func (Ctx *ConnCtx) ResetPassword(frame Frame) {
	if Ctx.IsAuthorized() {
		return
	}
	// generate bad ass password
	var user User
	var email string
	frame.BindJSON(&email)

	Ctx.Db.Where("email = ?", email).First(&user)
	if user.ID == 0 {
		Ctx.Pool.Write(Ctx.Ws, NewFrame(ResetPasswordFrame, "missing", frame.RequestID))
		return
	}
	newPassword, err := password.Generate(16, 4, 12, false, false)
	if err != nil {
		Ctx.Logger.WithError(err)
		Ctx.Pool.Write(Ctx.Ws, NewFrame(ResetPasswordFrame, "error", frame.RequestID))
		return
	}
	user.SetPassword(newPassword)
	Ctx.Db.Save(&user)
	mailer.SendPasswordReset(Ctx.Cfg, user, newPassword)
	Ctx.Pool.Write(Ctx.Ws, NewFrame(ResetPasswordFrame, "ok", frame.RequestID))
}

// UserInfo returns information about the user
func (Ctx *ConnCtx) UserInfo(frame Frame) {
	var request UserInfoRequest
	var user User
	frame.BindJSON(&request)
	Ctx.Db.Where("id = ? OR slug = ?", request.Slug, request.Slug).First(&user)
	if user.ID > 0 {
		Ctx.Pool.Write(Ctx.Ws, NewFrame(UserInfoFrame, user, frame.RequestID))
	}
}

// IsAuthorized checks if connection is authorized
func (Ctx *ConnCtx) IsAuthorized() bool {
	return Ctx.User != nil && Ctx.User.ID > 0
}

// UserUpdate saves user configuration
func (Ctx *ConnCtx) UserUpdate(frame Frame) {
	if Ctx.IsAuthorized() {
		var user User
		//json.Unmarshal([]byte(frame.Data), &user)
		frame.BindJSON(&user)
		if Ctx.User.ID == user.ID {
			Ctx.Db.First(Ctx.User)
			Ctx.User.Email = user.Email
			Ctx.User.Name = user.Name
			Ctx.User.Gender = user.Gender
			Ctx.Db.Save(Ctx.User)
			Ctx.respondWithToken(*Ctx.User)
		}
	}
}

// UserList returns users information
func (Ctx *ConnCtx) UserList(frame Frame) {
	var request UsersSearchRequest
	var users []User
	frame.BindJSON(&request)
	a := Ctx.Db.Where(
		"name LIKE ? OR slug = ? OR email LIKE ? ",
		request.Term+"%",
		request.Term,
		"%"+request.Term+"%",
	)
	if len(request.ExcludeUsers) > 0 {
		a = a.Where(" id NOT IN ( ? )", request.ExcludeUsers)
	}
	a.Limit(10).Find(&users)
	if users == nil {
		return
	}
	for _, u := range users {
		u = u.SafePluck()
	}
	Ctx.Pool.Write(Ctx.Ws, NewFrame(UserListFrame, users, frame.RequestID))
}

// UserDelete removes the user
func (Ctx *ConnCtx) UserDelete(frame Frame) {
	var user User
	frame.BindJSON(&user)
	if Ctx.IsAuthorized() && user.ID == Ctx.User.ID {
		Ctx.Db.Delete(&user)
	}
}
