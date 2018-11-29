package main

import (
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/liverecord/lrs"
	"github.com/liverecord/lrs/common"
	"github.com/liverecord/lrs/handlers"
)

var db *gorm.DB
var cfg *lrs.Config
var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	logger.Out = os.Stdout
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var pool = lrs.NewConnectionPool()

func handleConnections(w http.ResponseWriter, r *http.Request) {

	logger.Debug("handleConnections")
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	jwt := r.FormValue("jwt")
	logger.WithFields(logrus.Fields{"JWT": jwt}).Info("Request")
	if err != nil {
		logger.WithError(err).Error("Cannot upgrade protocol")
		return
	}
	// Make sure we close the connection when the function returns
	defer pool.DropConnection(ws)

	// Register our new client
	pool.AddConnection(ws)

	ws.WriteJSON(lrs.NewFrame(lrs.PingFrame, " ", ""))

	// our registry
	var lr = handlers.ConnCtx{
		Db:     db,
		Cfg:    cfg,
		Logger: logger,
		Ws:     ws,
		Pool:   pool,
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
		var f lrs.Frame
		mt, reader, err := ws.NextReader()
		if err != nil {
			logger.WithError(err).Errorln("Unable to read socket data")
			pool.DropConnection(ws)
			break
		}
		switch mt {
		case websocket.TextMessage:
			err = json.NewDecoder(reader).Decode(&f)
			if err != nil {
				logger.WithError(err).Errorln("Unable to read the Frame")

				// we drop this connection because Frames must be parsable
				pool.DropConnection(ws)
				break
			}
			logger.Debugf("Frame: %v", f)

			// We use reflection to call methods
			// Method name must match Frame.Type
			lrv := reflect.ValueOf(&lr)
			frv := reflect.ValueOf(f)
			method := lrv.MethodByName(f.Type)
			if method.IsValid() &&
				method.Type().NumIn() == 1 &&
				method.Type().In(0).AssignableTo(reflect.TypeOf(lrs.Frame{})) {
				method.Call([]reflect.Value{frv})
			} else {
				lr.Logger.Errorf("method %s is invalid", f.Type)
			}

		case websocket.BinaryMessage:
			if lr.IsAuthorized() {
				lr.Uploader(reader)
			} else {
				lr.Logger.Errorln("Unauthorized upload from", ws.RemoteAddr())
			}
		case websocket.CloseMessage:
			pool.DropConnection(ws)
			break
		}
	}
}

func handleOauth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Location", "/")
}

func handleStaticRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Location", "/")
}

func migrate(db *gorm.DB) {
	db.AutoMigrate(&lrs.Config{})
	db.AutoMigrate(&lrs.User{})
	db.AutoMigrate(&lrs.Topic{})
	db.AutoMigrate(&lrs.TopicStatus{})
	db.AutoMigrate(&lrs.Comment{})
	db.AutoMigrate(&lrs.Category{})
	db.AutoMigrate(&lrs.SocialProfile{})
	db.AutoMigrate(&lrs.Role{})
	db.AutoMigrate(&lrs.CommentStatus{})
	db.AutoMigrate(&lrs.Device{})
	db.AutoMigrate(&lrs.Settings{})
	db.AutoMigrate(&lrs.Attachment{})
}

func main() {
	logger.Infof("LiveRecord version 0.0.1")
	dotfile := ".env"
	for i, v := range os.Args {
		if (v == "--config" || v == "-c") && i < len(os.Args) - 1 {
			dotfile = (os.Args)[i + 1]
			break
		}
	}

	var err error
	err = godotenv.Load(dotfile)

	if err != nil {
		lrs.InteractiveSetup(logger)
		err = godotenv.Load(dotfile)
		if err != nil {
			logger.Panicln("Setup failed. Please, create .env file with configuration.")
		}
	}

	// open db connection
	db, err = gorm.Open(
		"mysql",
		common.Env("MYSQL_DSN", "root:123@tcp(127.0.0.1:3306)/liveRecord?charset=utf8&parseTime=True"))

	if err != nil {
		logger.WithError(err).Panic("Can't connect to the database")
	}

	defer db.Close()
	if common.BoolEnv("DEBUG", false) {
		db.LogMode(true)
		db.Debug()
		logger.SetLevel(logrus.DebugLevel)
	}

	// configure web-server
	// http.HandleFunc("/", handleStaticRequest)
	fs := http.FileServer(http.Dir(common.Env("DOCUMENT_ROOT", "assets")))
	http.Handle("/files/", fs)
	http.Handle("/app-dist/", fs)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/oauth/", handleOauth)
	http.HandleFunc("/api/oauth/facebook/", handleOauth)

	migrate(db)
	cfg = lrs.NewConfig(db, logger)

	lrs.RegisterStaticHandlers(cfg, db)

	ticker := time.NewTicker(time.Second)
	go func() {
		for _ = range ticker.C {
			/*
				pool.Broadcast(lrs.NewFrame(lrs.PingFrame, "", ""))

				var comment lrs.Comment
				Db.
					Preload("User").
					Preload("Topic").
					Order(gorm.Expr("rand()")).
					First(&comment)
			*/
			// pool.Broadcast(lrs.NewFrame(lrs.CommentFrame, comment, ""))
		}
	}()

	addr := common.Env("LISTEN_ADDR", "127.0.0.1:8000")
	logger.Infof("Listening on %s", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logger.WithError(err).Panic("Can't bind address & port")
	}
}
