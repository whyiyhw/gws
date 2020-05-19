package gws

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// 定义了监听 与 从 websocket 中读取数据 的结构体
type Conn struct {
	Conn *websocket.Conn

	AfterReadFunc   func(messageType int, r io.Reader)
	BeforeCloseFunc func()

	stopCh chan struct{}
}

// 往对应的连接中写入 UTF8 字符
// 成功返回对应 byte 的长度
// 失败返回 err
func (c *Conn) Write(p []byte) (n int, err error) {
	select {
	case <-c.stopCh:
		return 0, errors.New("连接已关闭, 写入失败")
	default:
		err = c.Conn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}
}

// Listen 监听 websocket 连接.  接收数据直到连接被关闭
func (c *Conn) Listen() {
	c.Conn.SetCloseHandler(func(code int, text string) error {
		if c.BeforeCloseFunc != nil {
			c.BeforeCloseFunc()
		}

		if err := c.close(); err != nil {
			log.Println(err)
		}

		message := websocket.FormatCloseMessage(code, "")
		_ = c.Conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
		return nil
	})

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

// close 主动关闭连接
func (c *Conn) close() error {
	select {
	case <-c.stopCh:
		return errors.New("连接已关闭")
	default:
		// 不允许主动调用 Close 方法 通过 conn.Close 调用 再通过回调函数进行调用
		err := c.Conn.Close()
		close(c.stopCh)
		return err
	}
}

// NewConn 新增对应连接
func NewConn(conn *websocket.Conn) *Conn {
	return &Conn{
		Conn:   conn,
		stopCh: make(chan struct{}),
	}
}
