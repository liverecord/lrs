package server

import (
	"github.com/gorilla/websocket"
)

type SocketConnectionsMap map[*websocket.Conn]*User
type UserConnectionsMap map[uint]map[*websocket.Conn]bool

type ConnectionPool struct {
	Sockets SocketConnectionsMap
	Users   UserConnectionsMap
}

func NewConnectionPool() *ConnectionPool {
	var pool ConnectionPool
	pool.Sockets = make(SocketConnectionsMap)
	pool.Users = make(UserConnectionsMap)
	return &pool
}

func (pool *ConnectionPool) AddConnection(conn *websocket.Conn) {
	pool.Sockets[conn] = nil
}

func (pool *ConnectionPool) Authenticate(conn *websocket.Conn, user *User) {
	pool.Sockets[conn] = user
	if _, ok := pool.Users[user.ID]; !ok {
		pool.Users[user.ID] = make(map[*websocket.Conn]bool)
	}
	pool.Users[user.ID][conn] = true
}

func (pool *ConnectionPool) Logout(user *User) {
	for conn := range pool.Users[user.ID] {
		pool.Sockets[conn] = nil
		delete(pool.Users[pool.Sockets[conn].ID], conn)
	}
}

func (pool *ConnectionPool) DropConnection(conn *websocket.Conn) {
	if pool.Sockets[conn] != nil {
		delete(pool.Users[pool.Sockets[conn].ID], conn)
	}
	delete(pool.Sockets, conn)
}

func (pool *ConnectionPool) Broadcast(frame Frame) {
	for conn := range pool.Sockets {
		err := conn.WriteJSON(frame)
		if err != nil {
			pool.DropConnection(conn)
		}
	}
}

func (pool *ConnectionPool) Send(to *User, frame Frame) {
	if _, ok := pool.Users[to.ID]; ok {
		for conn := range pool.Users[to.ID] {
			err := conn.WriteJSON(frame)
			if err != nil {
				pool.DropConnection(conn)
			}
		}
	}
}
