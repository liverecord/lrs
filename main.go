package main

import (
	"crypto/rand"

	"io"
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	. "github.com/liverecord/server/common"
	. "github.com/liverecord/server/common/common"
	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/handlers"
	"github.com/liverecord/server/model"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/liverecord/server/router"
)

var Db *gorm.DB
var Cfg *ServerConfig
var logger = logrus.New()

var frameRouter = router.NewFrame()

type Message struct {
	Email   string `json:"email"`
	Message string `json:"password"`
}

type LrClient struct {
	Conn *websocket.Conn
	User *model.User
}

var clients = make(handlers.SocketClientsMap) // connected clients
var broadcast = make(chan Message)            // broadcast channel
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

		// Register our new client
		clients[ws] = true

		w := map[string]interface{}{"type": 0, "connected": true}
		ws.WriteJSON(w)

		// our registry
		var lr = handlers.AppContext{
			Db:      Db,
			Cfg:     Cfg,
			Logger:  logger,
			Ws:      ws,
			Clients: &clients,
		}

		if len(jwt) > 0 {
			handlers.AuthorizeJWT(&lr, jwt)
		}

		for {
			// var msg Message
			var frame Frame
			// Read in a new message as JSON and map it to a Frame object
			err := ws.ReadJSON(&frame)

			//log.Printf("read error: %v, message: %d bytes %s", err, messageType, p)
			if err != nil {
				log.Print(err)
				delete(clients, ws)
				break
			}

			logger.Debugf("Frame: %+v", frame)
			f, err := frameRouter.Process(&lr, frame)
			if err != nil {
				lr.Logger.WithError(err).Error()
			} else if f.Type != "" {
				lr.Ws.WriteJSON(f)
			}
		}
	}
}
func handleOauth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Location", "/")
}

func handleBroadcastMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)

			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
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
		Env("MYSQL_DSN", "root:123@tcp(127.0.0.1:3306)/liveRecord?charset=utf8&parseTime=True"))

	// configure web-server
	fs := http.FileServer(http.Dir(Env("DOCUMENT_ROOT", "public")))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/oauth/", handleOauth)
	http.HandleFunc("/api/oauth/facebook/", handleOauth)

	frameRouter.AddHandler(CategoryListFrame, handlers.CategoryList)
	frameRouter.AddHandler(TopicFrame, handlers.Topic)
	frameRouter.AddHandler(TopicListFrame, handlers.TopicList)
	frameRouter.AddHandler(TopicDeleteFrame, handlers.TopicDelete)
	frameRouter.AddHandler(TopicSaveFrame, handlers.TopicSave)
	frameRouter.AddHandler(UserInfoFrame, handlers.UserInfo)
	frameRouter.AddHandler(UserUpdateFrame, handlers.UserUpdate)
	frameRouter.AddHandler(UserListFrame, handlers.UserList)
	frameRouter.AddHandler(UserDeleteFrame, handlers.UserDelete)
	frameRouter.AddHandler(AuthFrame, handlers.Auth)

	go handleBroadcastMessages()

	if err == nil {
		defer Db.Close()
		if BoolEnv("DEBUG", false) {
			Db.LogMode(true)
			Db.Debug()
			logger.SetLevel(logrus.DebugLevel)
		}
		Db.AutoMigrate(&ServerConfig{})
		Db.AutoMigrate(&model.User{})
		Db.AutoMigrate(&model.Topic{})
		Db.AutoMigrate(&model.Comment{})
		Db.AutoMigrate(&model.Category{})
		Db.AutoMigrate(&model.SocialProfile{})
		Db.AutoMigrate(&model.Role{})
		Db.AutoMigrate(&model.CommentStatus{})
		Db.AutoMigrate(&model.Attachment{})

		var configRecord ServerConfig
		Db.First(&configRecord)

		if configRecord.ID == 0 {
			configRecord.JwtSignature = make([]byte, 256)
			_, err = io.ReadFull(rand.Reader, configRecord.JwtSignature)
			logger.WithError(err)
			Db.Save(&configRecord)
		}

		Cfg = &configRecord

		addr := Env("LISTEN_ADDR", "127.0.0.1:8000")
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
