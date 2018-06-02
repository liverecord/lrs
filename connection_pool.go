package lrs

import (
	"github.com/gorilla/websocket"
)

// SocketStateMap type
type SocketStateMap map[*websocket.Conn]bool

// SocketConnectionsMap type
type SocketConnectionsMap map[*websocket.Conn]*User

// UserConnectionsMap type
type UserConnectionsMap map[uint64]SocketStateMap

// ConnectionPool is intended to keep track of all connections
type ConnectionPool struct {
	Sockets SocketConnectionsMap
	Users   UserConnectionsMap
	outbox  map[*websocket.Conn]chan Frame
}

// NewConnectionPool is factory to create new pool of connections
func NewConnectionPool() *ConnectionPool {
	var pool ConnectionPool
	pool.Sockets = make(SocketConnectionsMap)
	pool.Users = make(UserConnectionsMap)
	pool.outbox = make(map[*websocket.Conn]chan Frame)
	return &pool
}

// AddConnection adds connection to the pool
func (pool *ConnectionPool) AddConnection(conn *websocket.Conn) {
	pool.Sockets[conn] = nil
	pool.outbox[conn] = make(chan Frame)
	go pool.dispatch(conn)
}

// Authenticate binds user identifier to a connection
func (pool *ConnectionPool) Authenticate(conn *websocket.Conn, user *User) {
	pool.Sockets[conn] = user
	if _, ok := pool.Users[user.ID]; !ok {
		pool.Users[user.ID] = make(SocketStateMap)
	}
	pool.Users[user.ID][conn] = true
}

// Logout disaccosiates users and connections
func (pool *ConnectionPool) Logout(user *User) {
	for conn := range pool.Users[user.ID] {
		pool.Sockets[conn] = nil
		delete(pool.Users[pool.Sockets[conn].ID], conn)
	}
}

// DropConnection closes connection
func (pool *ConnectionPool) DropConnection(conn *websocket.Conn) {
	if pool.Sockets[conn] != nil {
		delete(pool.Users[pool.Sockets[conn].ID], conn)
	}
	if _, ok := pool.outbox[conn]; ok {
		close(pool.outbox[conn])
		delete(pool.outbox, conn)
	}
	delete(pool.Sockets, conn)
	conn.Close()
}

// Broadcast sends a frame to all connections
func (pool *ConnectionPool) Broadcast(frame Frame) {
	for conn := range pool.Sockets {
		pool.Write(conn, frame)
	}
}

// Write sends a frame to a specific connection
func (pool *ConnectionPool) Write(conn *websocket.Conn, frame Frame) {
	if _, ok := pool.outbox[conn]; ok {
		pool.outbox[conn] <- frame
	}
}

// Send delivers a frame to all user's connections
func (pool *ConnectionPool) Send(to *User, frame Frame) {
	if _, ok := pool.Users[to.ID]; ok {
		for conn := range pool.Users[to.ID] {
			pool.Write(conn, frame)
		}
	}
}

// actually sends frames
func (pool *ConnectionPool) dispatch(c *websocket.Conn) {
	for f := range pool.outbox[c] {
		err := c.WriteJSON(&f)
		if err != nil {
			pool.DropConnection(c)
		}
	}
}
