package server

import (
	"github.com/gorilla/websocket"
)

type SocketStateMap map[*websocket.Conn]bool
type SocketConnectionsMap map[*websocket.Conn]*User
type UserConnectionsMap map[uint]SocketStateMap

type ConnectionPool struct {
	Sockets SocketConnectionsMap
	Users   UserConnectionsMap
	outbox  map[*websocket.Conn]chan Frame
}

func NewConnectionPool() *ConnectionPool {
	var pool ConnectionPool
	pool.Sockets = make(SocketConnectionsMap)
	pool.Users = make(UserConnectionsMap)
	pool.outbox = make(map[*websocket.Conn]chan Frame)
	return &pool
}

func (pool *ConnectionPool) AddConnection(conn *websocket.Conn) {
	pool.Sockets[conn] = nil
	pool.outbox[conn] = make(chan Frame)
	go pool.dispatch(conn)
}

func (pool *ConnectionPool) Authenticate(conn *websocket.Conn, user *User) {
	pool.Sockets[conn] = user
	if _, ok := pool.Users[user.ID]; !ok {
		pool.Users[user.ID] = make(SocketStateMap)
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
	if _, ok := pool.outbox[conn]; ok {
		close(pool.outbox[conn])
		delete(pool.outbox, conn)
	}
	delete(pool.Sockets, conn)
	conn.Close()
}

func (pool *ConnectionPool) Broadcast(frame Frame) {
	for conn := range pool.Sockets {
		pool.Write(conn, frame)
	}
}

func (pool *ConnectionPool) Write(conn *websocket.Conn, frame Frame) {
	if _, ok := pool.outbox[conn]; ok {
		pool.outbox[conn] <- frame
	}
}

func (pool *ConnectionPool) Send(to *User, frame Frame) {
	if _, ok := pool.Users[to.ID]; ok {
		for conn := range pool.Users[to.ID] {
			pool.Write(conn, frame)
		}
	}
}

func (pool *ConnectionPool) dispatch(c *websocket.Conn) {
	for f := range pool.outbox[c] {
		// f, _ := <-pool.outbox[c]
		// time.Sleep(100 * time.Millisecond) // testing network delays

		err := c.WriteJSON(&f)
		if err != nil {
			pool.DropConnection(c)
		}
	}
}
