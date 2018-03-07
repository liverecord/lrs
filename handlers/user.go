package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"

	"errors"

	. "github.com/liverecord/server/common/common"
	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/model"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginData struct {
	Jwt  string     `json:"jwt"`
	User model.User `json:"user"`
}

type UserInfoRequest struct {
	Slug string `json:"slug"`
}

type UsersSearchRequest struct {
	Term         string `json:"term"`
	ExcludeUsers []uint `json:"exclude"`
}

func Auth(ctx *AppContext, frame Frame) (Frame, error) {
	var authData model.UserAuthData
	frame.BindJSON(&authData)
	ctx.Logger.Debugf("AuthData: %v", authData)
	var user model.User
	authData.Email = strings.ToLower(authData.Email)
	ctx.Db.Where("email = ?", authData.Email).First(&user)

	if ctx.IsAuthorized() {
		err := bcrypt.CompareHashAndPassword(
			S2BA(user.Password),
			S2BA(authData.Password),
		)

		if err == nil {
			// we are cool, password is correct
			ctx.User = &user
			return respondWithToken(ctx, user), nil
		}

		ctx.Logger.WithError(err).Errorf("Cannot authorize user %s %v", user.Email, err)
		return Frame{Type: AuthErrorFrame, Data: "PasswordMismatch"}, nil
	}

	user.Email = authData.Email
	user.Name = user.MakeNameFromEmail()
	user.Roles = []model.Role{}
	user.MakeSlug()
	user.SetPassword(authData.Password)
	user.Picture = user.MakeGravatarPicture()
	ctx.Db.Save(&user)
	ctx.User = &user

	return respondWithToken(ctx, user), nil
}

func respondWithToken(ctx *AppContext, user model.User) Frame {
	uld, err := generateToken(user, ctx.Cfg.JwtSignature)
	if err != nil {
		ctx.Logger.WithError(err).Printf("Cannot generate token %v", err)
		return Frame{}
	}

	userData, err := json.Marshal(uld)
	if err != nil {
		ctx.Logger.WithError(err).Error("Cannot marshall user data")
		return Frame{}
	}

	return Frame{Type: AuthFrame, Data: string(userData)}
}

func generateToken(user model.User, signature []byte) (UserLoginData, error) {
	// Create the token
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"uid":  user.ID,
			"hash": user.Hash,
			"exp":  time.Now().Add(time.Hour * 24 * 30).Unix(),
		},
	)

	// Sign and get the complete encoded token as a string
	var uld UserLoginData
	var err error
	uld.User = user
	uld.Jwt, err = token.SignedString(signature)

	return uld, err
}

func AuthorizeJWT(ctx *AppContext, tokenString string) (Frame, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return ctx.Cfg.JwtSignature, nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		ctx.Logger.Error(err)
		return Frame{}, nil
	}

	fmt.Println(claims["uid"], claims["exp"])
	ctx.Logger.WithFields(logrus.Fields(claims)).Debug("Debugged tokens")
	var user model.User
	ctx.Db.Where("id = ? AND hash = ?", claims["uid"], claims["hash"]).First(&user)

	if user.ID != 0 {
		ctx.User = &user
		return respondWithToken(ctx, user), nil
	}

	return Frame{}, nil
}

func UserInfo(ctx *AppContext, frame Frame) (Frame, error) {
	var request UserInfoRequest
	var user model.User

	frame.BindJSON(&request)
	ctx.Db.Where("id = ? OR slug = ?", request.Slug, request.Slug).First(&user)
	if user.ID == 0 {
		return Frame{}, fmt.Errorf("could not find user %s", request.Slug)
	}

	userData, err := json.Marshal(user)
	if err != nil {
		return Frame{}, err
	}

	return Frame{Type: UserInfoFrame, Data: string(userData)}, nil
}

func UserUpdate(ctx *AppContext, frame Frame) (Frame, error) {
	if !ctx.IsAuthorized() {
		return Frame{}, errors.New("we can update only authorized user")
	}

	var user model.User

	frame.BindJSON(&user)
	if ctx.User.ID != user.ID {
		return Frame{}, errors.New("user can not update another user")
	}

	ctx.Db.First(ctx.User)
	ctx.User.Email = user.Email
	ctx.User.Name = user.Name
	ctx.User.Gender = user.Gender
	ctx.Db.Save(ctx.User)

	return respondWithToken(ctx, *ctx.User), nil
}

func UserList(ctx *AppContext, frame Frame) (Frame, error) {
	var request UsersSearchRequest
	var users []model.User
	frame.BindJSON(&request)

	a := ctx.Db.Where(
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
		return Frame{}, errors.New("users was not found")
	}

	userData, err := json.Marshal(users)
	if err != nil {
		return Frame{}, err
	}

	return Frame{Type: UserListFrame, Data: string(userData)}, nil
}

func UserDelete(ctx *AppContext, frame Frame) (Frame, error) {
	var user model.User
	frame.BindJSON(&user)
	ctx.Db.Delete(&user)

	return Frame{}, nil
}
