package gws

import (
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Conn wraps websocket.Conn with Conn. 定义了监听与 从中读取数据
type Conn struct {
	Conn *websocket.Conn

	AfterReadFunc   func(messageType int, r io.Reader)
	BeforeCloseFunc func()
	MessageFunc     func()

	once   sync.Once
	id     string
	stopCh chan struct{}
}

// Write write p to the websocket connection. The error returned will always
// be nil if success.
func (c *Conn) Write(p []byte) (n int, err error) {
	select {
	case <-c.stopCh:
		return 0, errors.New("Conn is closed, can't be written")
	default:
		err = c.Conn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}
}

// GetID returns the id generated using UUID algorithm.
func (c *Conn) GetID() string {
	c.once.Do(func() {
		u := uuid.New()
		c.id = u.String()
	})

	return c.id
}

// Listen listens for receive data from websocket connection. It blocks
// until websocket connection is closed.
func (c *Conn) Listen() {
	c.Conn.SetCloseHandler(func(code int, text string) error {
		if c.BeforeCloseFunc != nil {
			c.BeforeCloseFunc()
		}

		if err := c.Close(); err != nil {
			log.Println(err)
		}

		message := websocket.FormatCloseMessage(code, "")
		_ = c.Conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
		return nil
	})

	// Keeps reading from Conn util get error.
ReadLoop:
	for {
		select {
		case <-c.stopCh:
			break ReadLoop
		default:
			messageType, r, err := c.Conn.NextReader()
			if err != nil {
				// TODO: 需要去处理读取数据错误
				// 可能会出现 websocket: close 1001 (going away)
				//fmt.Printf("read msg err %s", err.Error())
				break ReadLoop
			}

			if c.AfterReadFunc != nil {
				c.AfterReadFunc(messageType, r)
			}
		}
	}
}

// Close close the connection.
func (c *Conn) Close() error {
	select {
	case <-c.stopCh:
		return errors.New("Conn already been closed")
	default:
		_ = c.Conn.Close()
		close(c.stopCh)
		return nil
	}
}

// NewConn wraps conn.
func NewConn(conn *websocket.Conn) *Conn {
	return &Conn{
		Conn:   conn,
		stopCh: make(chan struct{}),
	}
}
