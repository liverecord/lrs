package main

import (
	"crypto/rand"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/liverecord/server"
	"github.com/liverecord/server/common"
	"github.com/liverecord/server/handlers"
)

var Db *gorm.DB
var Cfg *server.Config
var logger = logrus.New()

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var pool = server.NewConnectionPool()

func handleConnections(w http.ResponseWriter, r *http.Request) {

	logger.Debug("handleConnections")
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	jwt := r.FormValue("jwt")
	logger.WithFields(logrus.Fields{"JWT": jwt}).Info("Request")
	if err != nil {
		logger.WithError(err).Error("Cannot upgrade protocol")
	} else {
		// Make sure we close the connection when the function returns
		defer ws.Close()
		defer pool.DropConnection(ws)

		// Register our new client
		pool.AddConnection(ws)

		ws.WriteJSON(server.NewFrame(server.PingFrame, " ", ""))

		// our registry
		var lr = handlers.AppContext{
			Db:      Db,
			Cfg:     Cfg,
			Logger:  logger,
			Ws:      ws,
			Pool:    pool,
		}

		if len(jwt) > 0 {
			lr.AuthorizeJWT(jwt)
			if lr.IsAuthorized() {
				pool.Authenticate(ws, lr.User)
			}
		}

		for {
			var f server.Frame
			err := ws.ReadJSON(&f)
			if err != nil {
				logger.WithError(err).Errorln("Unable to read request")
				pool.DropConnection(ws)
				break
			} else {
				logger.Debugf("Frame: %v", f)
				// We use reflection to call methods
				// Method name must match Frame.Type
				lrv := reflect.ValueOf(&lr)
				frv := reflect.ValueOf(f)
				method := lrv.MethodByName(f.Type)
				if method.IsValid() &&
					method.Type().NumIn() == 1 &&
					method.Type().In(0).AssignableTo(reflect.TypeOf(server.Frame{})) {
					method.Call([]reflect.Value{frv})
				} else {
					lr.Logger.Errorf("method %s is invalid", f.Type)
				}
			}
		}
	}
}

func handleOauth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Location", "/")
}

func main() {
	var err error
	err = godotenv.Load()

	logger.Out = os.Stdout

	if err != nil {
		logger.Fatal("Error loading .env file")
	}
	// open db connection
	Db, err = gorm.Open(
		"mysql",
		common.Env("MYSQL_DSN", "root:123@tcp(127.0.0.1:3306)/liveRecord?charset=utf8&parseTime=True"))

	// configure web-server
	fs := http.FileServer(http.Dir(common.Env("DOCUMENT_ROOT", "public")))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/oauth/", handleOauth)
	http.HandleFunc("/api/oauth/facebook/", handleOauth)

	if err == nil {
		defer Db.Close()
		if common.BoolEnv("DEBUG", false) {
			Db.LogMode(true)
			Db.Debug()
			logger.SetLevel(logrus.DebugLevel)
		}
		Db.AutoMigrate(&server.Config{})
		Db.AutoMigrate(&server.User{})
		Db.AutoMigrate(&server.Topic{})
		Db.AutoMigrate(&server.Comment{})
		Db.AutoMigrate(&server.Category{})
		Db.AutoMigrate(&server.SocialProfile{})
		Db.AutoMigrate(&server.Role{})
		Db.AutoMigrate(&server.CommentStatus{})
		Db.AutoMigrate(&server.Attachment{})

		var config server.Config
		Db.First(&config)

		if config.ID == 0 {
			config.JwtSignature = make([]byte, 256)
			_, err = io.ReadFull(rand.Reader, config.JwtSignature)
			logger.WithError(err)
			Db.Save(&config)
		}

		Cfg = &config

		addr := common.Env("LISTEN_ADDR", "127.0.0.1:8000")
		logger.Printf("Listening on %s", addr)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			logger.WithError(err).Fatal("Can't bind address & port")
		}
	} else {
		logger.WithError(err).Fatal("Can't bind address & port")
		logger.Panic(err)
	}
}
