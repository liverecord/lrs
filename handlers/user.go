package user

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"

	. "github.com/liverecord/server/common/common"
	. "github.com/liverecord/server/common/frame"
	. "github.com/liverecord/server/model"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginData struct {
	Jwt  string `json:"jwt"`
	User User   `json:"user"`
}

func LoginHandler(cfg *ServerConfig, ws *websocket.Conn, Db *gorm.DB, frame Frame) {
	var authData UserAuthData

	json.Unmarshal([]byte(frame.Data), &authData)
	log.Printf("AuthData: %v", authData)
	var user User
	Db.Where("email = ?", authData.Email).First(&user)
	if user.ID != 0 {
		err := bcrypt.CompareHashAndPassword(
			S2BA(user.Password),
			S2BA(authData.Password))
		if err == nil {
			// we are cool, password is correct
			respondWithToken(user, cfg, ws)
		} else {
			ws.WriteJSON(Frame{Type: AuthErrorFrame, Data: "PasswordMismatch"})
			log.Printf("Cannot authorize user %s %v", user.Email, err)
		}

	} else {
		user.Email = authData.Email
		user.Name = user.MakeNameFromEmail()
		user.MakeSlug()
		user.SetPassword(authData.Password)
		user.Picture = user.MakeGravatarPicture()
		Db.Save(&user)
		respondWithToken(user, cfg, ws)
	}
}

func respondWithToken(user User, cfg *ServerConfig, ws *websocket.Conn) {
	uld, err := generateToken(user, cfg)
	if err == nil {
		userData, err := json.Marshal(uld)
		if err == nil {
			ws.WriteJSON(Frame{Type: AuthFrame, Data: string(userData)})
		} else {
			log.Printf("Cannot marshall user data %v", err)
		}
	} else {
		log.Printf("Cannot generate token %v", err)
	}
}

func generateToken(user User, cfg *ServerConfig) (UserLoginData, error) {
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
	uld.Jwt, err = token.SignedString(cfg.JwtSignature)
	return uld, err
}

func AuthorizeJWT(tokenString string, cfg *ServerConfig, ws *websocket.Conn, Db *gorm.DB) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return cfg.JwtSignature, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["uid"], claims["exp"])
		logrus.WithFields(logrus.Fields(claims)).Debug("Debugged tokens")
		var user User
		Db.Where("id = ? AND hash = ?", claims["uid"], claims["hash"]).First(&user)
		if user.ID != 0 {
			respondWithToken(user, cfg, ws)
		}

	} else {
		logrus.Error(err)
	}

}

func UserUpdateHandler(cfg *ServerConfig, ws *websocket.Conn, Db *gorm.DB, frame Frame) {
}

func JWTAuthHandler(cfg *ServerConfig, ws *websocket.Conn, Db *gorm.DB, frame Frame) {
}
