package handlers

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	"github.com/gorilla/websocket"
	"github.com/liverecord/lrs"
)

type AppContext struct {
	Cfg    *lrs.Config
	Db     *gorm.DB
	Logger *logrus.Logger
	Ws     *websocket.Conn
	User   *lrs.User
	Pool   *lrs.ConnectionPool
	File   *File
}

type ConnectionContext struct {
	App  *AppContext
	Ws   *websocket.Conn
	User *lrs.User
}
