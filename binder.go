package gws

import (
	"errors"
	"fmt"
	"sync"
)

// binder 为 生成与解绑 id 与 conn 之间关系 的结构体
type binder struct {
	mu sync.RWMutex

	// userID 跟 连接的 map
	userID2ConnMap map[int]*Conn
}

// Bind 绑定 userID 跟 对应的连接
func (b *binder) Bind(userID int, conn *Conn) error {

	if conn == nil {
		return errors.New("连接不能为空")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.userID2ConnMap[userID]; ok {
		// 该userID 对应的连接已存在
		return nil
	} else {
		b.userID2ConnMap[userID] = conn
	}

	return nil
}

// Unbind 解绑 userID 跟 对应的连接
func (b *binder) Unbind(conn *Conn) error {
	if conn == nil {
		return errors.New("连接不能为空")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for userID, c := range b.userID2ConnMap {
		if c == conn {
			delete(b.userID2ConnMap, userID)
			return nil
		}
	}

	return fmt.Errorf("该连接不在连接 map 中 连接ID为  %s", conn.GetID())
}

// FindByID 从 map 中 找到对应的连接
func (b *binder) FindByID(userID int) (c *Conn, err error) {

	b.mu.RLock()
	defer b.mu.RUnlock()

	if c, ok := b.userID2ConnMap[userID]; ok {
		return c, nil
	}

	err = errors.New("该链接不存在，或已失效")
	return

}

//FindIDByConn 通过连接去找 ID
func (b *binder) FindIDByConn(conn *Conn) (userID int, err error) {

	b.mu.RLock()
	defer b.mu.RUnlock()

	for userID, c := range b.userID2ConnMap {
		if c == conn {
			return userID, nil
		}
	}

	err = errors.New("该链接不存在，或已失效")
	return

}
