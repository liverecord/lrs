package handlers

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	"github.com/gorilla/websocket"
	"github.com/liverecord/lrs"
)

// AppContext defines current application context
type AppContext struct {
	Cfg    *lrs.Config
	Db     *gorm.DB
	Logger *logrus.Logger
	Ws     *websocket.Conn
	User   *lrs.User
	Pool   *lrs.ConnectionPool
	File   *File
}

// ConnectionContext defines current application context
type ConnectionContext struct {
	App  *AppContext
	Ws   *websocket.Conn
	User *lrs.User
}
