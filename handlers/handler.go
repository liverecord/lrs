package handlers

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/liverecord/server/model"

	"github.com/gorilla/websocket"

	"github.com/liverecord/server/common/common"
)

type SocketClientsMap map[*websocket.Conn]bool

type AppContext struct {
	Cfg     *common.ServerConfig
	Db      *gorm.DB
	Logger  *logrus.Logger
	Ws      *websocket.Conn
	User    *model.User
	Clients *SocketClientsMap
}
