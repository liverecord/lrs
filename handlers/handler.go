package handlers

import (
	"github.com/sirupsen/logrus"
	"github.com/jinzhu/gorm"

	"github.com/gorilla/websocket"
	"github.com/liverecord/lrs"
)

// ConnCtx defines current application context
type ConnCtx struct {
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
	App  *ConnCtx
	Ws   *websocket.Conn
	User *lrs.User
}
