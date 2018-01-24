package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"

	. "github.com/liverecord/server/common/common"
	. "github.com/liverecord/server/common/frame"
	. "github.com/liverecord/server/model"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginData struct {
	Jwt  string `json:"jwt"`
	User User   `json:"user"`
}

type UserInfoRequest struct {
	Slug string `json:"slug"`
}

type UsersSearchRequest struct {
	Term         string `json:"term"`
	ExcludeUsers []uint `json:"exclude"`
}

func (Ctx *AppContext) Login(frame Frame) {
	var authData UserAuthData
	frame.BindJSON(&authData)
	Ctx.Logger.Debugf("AuthData: %v", authData)
	var user User
	authData.Email = strings.ToLower(authData.Email)
	Ctx.Db.Where("email = ?", authData.Email).First(&user)
	if Ctx.IsAuthorized() {
		err := bcrypt.CompareHashAndPassword(
			S2BA(user.Password),
			S2BA(authData.Password))
		if err == nil {
			// we are cool, password is correct
			Ctx.respondWithToken(user)
			Ctx.User = &user
		} else {
			Ctx.Ws.WriteJSON(Frame{Type: AuthErrorFrame, Data: "PasswordMismatch"})
			Ctx.Logger.WithError(err).Errorf("Cannot authorize user %s %v", user.Email, err)
		}

	} else {
		user.Email = authData.Email
		user.Name = user.MakeNameFromEmail()
		user.Roles = []Role{}
		user.MakeSlug()
		user.SetPassword(authData.Password)
		user.Picture = user.MakeGravatarPicture()
		Ctx.Db.Save(&user)
		Ctx.User = &user
		Ctx.respondWithToken(user)
	}
}

func (Ctx *AppContext) respondWithToken(user User) {
	uld, err := Ctx.generateToken(user)
	if err == nil {
		userData, err := json.Marshal(uld)
		if err == nil {
			Ctx.Ws.WriteJSON(Frame{Type: AuthFrame, Data: string(userData)})
		} else {
			Ctx.Logger.WithError(err).Error("Cannot marshall user data")
		}
	} else {
		Ctx.Logger.WithError(err).Printf("Cannot generate token %v", err)
	}
}

func (Ctx *AppContext) generateToken(user User) (UserLoginData, error) {
	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
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

func (Ctx *AppContext) AuthorizeJWT(tokenString string) {

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

	} else {
		Ctx.Logger.Error(err)
	}

}

func (Ctx *AppContext) UserInfo(frame Frame) {
	var request UserInfoRequest
	var user User
	frame.BindJSON(&request)
	Ctx.Db.Where("id = ? OR slug = ?", request.Slug, request.Slug).First(&user)
	if user.ID > 0 {
		userData, err := json.Marshal(user)
		if err == nil {
			Ctx.Ws.WriteJSON(Frame{Type: UserInfoFrame, Data: string(userData)})
		} else {
			Ctx.Logger.WithError(err)
		}
	}
}

func (Ctx *AppContext) broadcast(frame Frame) {

}

func (Ctx *AppContext) IsAuthorized() bool {
	return Ctx.User != nil && Ctx.User.ID > 0
}

func (Ctx *AppContext) UserUpdate(frame Frame) {
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

func (Ctx *AppContext) UsersList(frame Frame) {
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
	if users != nil {
		userData, err := json.Marshal(users)
		if err == nil {
			Ctx.Ws.WriteJSON(Frame{Type: UserListFrame, Data: string(userData)})
		} else {
			Ctx.Logger.WithError(err)
		}
	}
}

func (Ctx *AppContext) UserDelete(frame Frame) {
	var user User
	frame.BindJSON(&user)
	Ctx.Db.Delete(&user)
}
