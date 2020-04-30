package gws

import (
	"errors"
	"fmt"
	"sync"
)

// eventConn wraps Conn with a specified event type.
type eventConn struct {
	Event string
	Conn  *Conn
}

// binder is defined to store the relation of userID and eventConn
type binder struct {
	mu sync.RWMutex

	// map stores key: userID and value of related slice of eventConn
	userID2ConnMap map[int]*Conn

}

// Bind binds userID with eConn specified by event. It fails if the
// return error is not nil.
func (b *binder) Bind(userID int, conn *Conn) error {


	if conn == nil {
		return errors.New("conn can't be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// map the eConn if it isn't be put.
	if _, ok := b.userID2ConnMap[userID]; ok {
		// 如果维持了超过 10240 个链接，那么对于重复 链接 怎么处理？
		return nil
	} else {
		b.userID2ConnMap[userID] = conn
	}

	return nil
}

// Unbind unbind and removes Conn if it's exist.
func (b *binder) Unbind(conn *Conn) error {
	if conn == nil {
		return errors.New("conn can't be empty")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for userID, c := range b.userID2ConnMap {
		if c == conn {
			delete(b.userID2ConnMap, userID)
			return nil
		}
	}

	return fmt.Errorf("can't find the conn of ID: %s", conn.GetID())
}

// FilterConn searches the conns related to userID, and filtered by
// event. The userID can't be empty. The event will be ignored if it's empty.
// All the conns related to the userID will be returned if the event is empty.
func (b *binder) FilterConn(userID int) (*Conn, error) {

	b.mu.RLock()
	defer b.mu.RUnlock()

	if eConns, ok := b.userID2ConnMap[userID]; ok {

		return eConns, nil
	}else{
		//TODO 需要去处理
		return eConns, nil
	}
}
