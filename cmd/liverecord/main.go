package main

import (
	"crypto/rand"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"

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
var logger *logrus.Logger

func init()  {
 	logger = logrus.New()
 	logger.Formatter = &logrus.TextFormatter{ForceColors:true}
	logger.Out = os.Stdout
}

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
		// The Magic Frame router
		//
		// Intention of this router serves simple purpose of providing easy way to develop
		// and extend this application
		// For example, you can build plugins with your methods and extend this app
		// The current implementation is a rough idea of self-declaring routing
		for {
			var f server.Frame
			err := ws.ReadJSON(&f)
			if err != nil {
				logger.WithError(err).Errorln("Unable to read request")
				// we drop this connection because Frames must be parsable
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

	if err != nil {
		logger.Panic("Error loading .env file")
	}

	// open db connection
	Db, err = gorm.Open(
		"mysql",
		common.Env("MYSQL_DSN", "root:123@tcp(127.0.0.1:3306)/liveRecord?charset=utf8&parseTime=True"))

	if err != nil {
		logger.WithError(err).Panic("Can't connect to the database")
	}

	defer Db.Close()
	if common.BoolEnv("DEBUG", false) {
		Db.LogMode(true)
		Db.Debug()
		logger.SetLevel(logrus.DebugLevel)
	}

	// configure web-server
	fs := http.FileServer(http.Dir(common.Env("DOCUMENT_ROOT", "assets")))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/oauth/", handleOauth)
	http.HandleFunc("/api/oauth/facebook/", handleOauth)

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
		// lets set this application with default parameters
		config.JwtSignature = make([]byte, 256)
		_, err = io.ReadFull(rand.Reader, config.JwtSignature)
		logger.WithError(err).Errorln("Unable to generate JWT Signature")
		Db.Save(&config)
	}

	Cfg = &config

	ticker := time.NewTicker(time.Second)

	go func() {
		for _ = range ticker.C {
			/*
			pool.Broadcast(server.NewFrame(server.PingFrame, "", ""))

			var comment server.Comment
			Db.
				Preload("User").
				Preload("Topic").
				Order(gorm.Expr("rand()")).
				First(&comment)
*/
			// pool.Broadcast(server.NewFrame(server.CommentFrame, comment, ""))
		}
	}()

	addr := common.Env("LISTEN_ADDR", "127.0.0.1:8000")
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logger.WithError(err).Panic("Can't bind address & port")
	}
	logger.Printf("Listening on %s", addr)


}
