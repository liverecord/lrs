package handlers

import (
	"encoding/json"
	"fmt"
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

func (Lr *LiveRecord) Login(frame Frame) {
	var authData UserAuthData

	json.Unmarshal([]byte(frame.Data), &authData)
	Lr.Logger.Debugf("AuthData: %v", authData)
	var user User
	Lr.Db.Where("email = ?", authData.Email).First(&user)
	if user.ID != 0 {
		err := bcrypt.CompareHashAndPassword(
			S2BA(user.Password),
			S2BA(authData.Password))
		if err == nil {
			// we are cool, password is correct
			Lr.respondWithToken(user)
			Lr.User = &user
		} else {
			Lr.Ws.WriteJSON(Frame{Type: AuthErrorFrame, Data: "PasswordMismatch"})
			Lr.Logger.WithError(err).Errorf("Cannot authorize user %s %v", user.Email, err)
		}

	} else {
		user.Email = authData.Email
		user.Name = user.MakeNameFromEmail()
		user.Roles = []Role{}
		user.MakeSlug()
		user.SetPassword(authData.Password)
		user.Picture = user.MakeGravatarPicture()
		Lr.Db.Save(&user)
		Lr.User = &user
		Lr.respondWithToken(user)
	}
}

func (Lr *LiveRecord) respondWithToken(user User) {
	uld, err := Lr.generateToken(user)
	if err == nil {
		userData, err := json.Marshal(uld)
		if err == nil {
			Lr.Ws.WriteJSON(Frame{Type: AuthFrame, Data: string(userData)})
		} else {
			Lr.Logger.WithError(err).Error("Cannot marshall user data")
		}
	} else {
		Lr.Logger.WithError(err).Printf("Cannot generate token %v", err)
	}
}

func (Lr *LiveRecord) generateToken(user User) (UserLoginData, error) {
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
	uld.Jwt, err = token.SignedString(Lr.Cfg.JwtSignature)
	return uld, err
}

func (Lr *LiveRecord) AuthorizeJWT(tokenString string) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return Lr.Cfg.JwtSignature, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["uid"], claims["exp"])
		Lr.Logger.WithFields(logrus.Fields(claims)).Debug("Debugged tokens")
		var user User
		Lr.Db.Where("id = ? AND hash = ?", claims["uid"], claims["hash"]).First(&user)
		if user.ID != 0 {
			Lr.respondWithToken(user)
		}

	} else {
		Lr.Logger.Error(err)
	}

}

func  (Lr *LiveRecord) UserInfo(frame Frame) {
	var request UserInfoRequest
	var user User
	json.Unmarshal([]byte(frame.Data), &request)
	Lr.Db.Where("id = ? OR slug = ?", request.Slug, request.Slug).First(&user)
	if user.ID > 0 {
		userData, err := json.Marshal(user)
		if err == nil {
			Lr.Ws.WriteJSON(Frame{Type: UserInfoFrame, Data: string(userData)})
		} else {
			logrus.WithError(err)
		}
	}
}

func (Lr *LiveRecord) UserUpdate(frame Frame) {
	if Lr.User == nil {

	} else {
		var user User
		json.Unmarshal([]byte(frame.Data), &user)
		if Lr.User.ID == user.ID {
			Lr.Db.First(Lr.User)
			Lr.User.Email = user.Email
			Lr.User.Name = user.Name
			Lr.User.Gender = user.Gender
			Lr.Db.Save(Lr.User)
			Lr.respondWithToken(*Lr.User)
		}
	}
}

func (Lr *LiveRecord) UserDelete(frame Frame) {
	var user User
	json.Unmarshal([]byte(frame.Data), &user)
	Lr.Db.Delete(&user)
}
