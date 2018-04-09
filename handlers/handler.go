package handlers

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/liverecord/server"

	"github.com/gorilla/websocket"
)

type AppContext struct {
	Cfg     *server.Config
	Db      *gorm.DB
	Logger  *logrus.Logger
	Ws      *websocket.Conn
	User    *server.User
	Pool 	*server.ConnectionPool
}

type ConnectionContext struct {
	App     *AppContext
	Ws      *websocket.Conn
	User    *server.User
}