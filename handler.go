package gws

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var UserID int

// websocketHandler defines to handle websocket upgrade request.
type websocketHandler struct {
	// upgrader 是需要去升级的请求
	upgrader *websocket.Upgrader

	// 绑定者 处理 websocket连接与客户端ID 之间的联系
	binder *binder

	mu sync.RWMutex

	// 连接事件
	onOpen func(conn *Conn, fd int)
	// 消息接受事件
	onMessage func(conn *Conn, fd int, message string, err error)
	// 连接关闭事件
	onClose func(conn *Conn, fd int)
}

// 首先尝试去升级连接为 websocket协议，如果 success, 长连接会一直保活下去
// 直到客户端发送关闭 信息，或者 服务端主动 drop 掉
func (wh *websocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConn, err := wh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer func() {
		_ = wsConn.Close()
	}()

	// 处理 Websocket 请求
	conn := NewConn(wsConn)

	if UserID > 1024000 {
		UserID = 1
	} else {
		UserID++
	}
	err = wh.binder.Bind(UserID, conn)
	if err != nil {
		fmt.Println(err.Error())
		return // 不再进行下一步操作
	}
	if wh.onOpen != nil {
		userID, err := wh.binder.FindIDByConn(conn)
		if err != nil {
			fmt.Println("open 事件", err.Error())
			return // 不再进行下一步操作
		}
		wh.onOpen(conn, userID)
	}
	conn.AfterReadFunc = func(messageType int, r io.Reader) {
		// 这里是 message 事件
		if wh.onMessage != nil {
			p, err := io.ReadAll(r)
			if err != nil {
				fmt.Println("message 读取data 失败", err.Error())
				return // 不再进行下一步操作
			}
			userID, err := wh.binder.FindIDByConn(conn)
			if err != nil {
				fmt.Println("message 事件", err.Error())
				return // 不再进行下一步操作
			}
			wh.onMessage(conn, userID, string(p), err)
		}
	}
	conn.BeforeCloseFunc = func() {
		// unbind 这里是 close  事件
		if wh.onClose != nil {
			// 需要根据 conn 获取 userID
			userID, err := wh.binder.FindIDByConn(conn)
			if err != nil {
				fmt.Println("close 事件 获取 userID 失败", err.Error())
				return // 不再进行下一步操作
			}
			wh.onClose(conn, userID)
		}
		err = wh.binder.Unbind(conn)
		if err != nil {
			fmt.Println("close 事件 解绑失败", err.Error())
			return // 不再进行下一步操作
		}
	}

	conn.Listen()
}

// closeConn 通过 userID 去解绑 conn
func (wh *websocketHandler) closeConn(userID int) (int, error) {
	conns, _ := wh.binder.FindByID(userID)

	if err := wh.binder.Unbind(conns); err != nil {
		log.Printf("conn unbind fail: %v", err)
		return 0, nil
	}

	return 1, nil
}

// HttpHandler 定义http处理 func
type HttpHandler struct {
	Path     string
	DealFunc http.HandlerFunc
}
