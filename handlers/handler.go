package handlers

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/liverecord/server"

	"github.com/gorilla/websocket"
)

//type SocketClientsMap map[*websocket.Conn]bool
//type SocketUsersMap map[*websocket.Conn]*server.User

type AppContext struct {
	Cfg     *server.Config
	Db      *gorm.DB
	Logger  *logrus.Logger
	Ws      *websocket.Conn
	User    *server.User
	Pool 	*server.ConnectionPool
}
