package gws

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var UserID int

// websocketHandler defines to handle websocket upgrade request.
type websocketHandler struct {
	// upgrader is used to upgrade request.
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

// RegisterMessage defines message struct client send after connect
// to the server.
type RegisterMessage struct {
	Token string
	Event string
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

	// handle Websocket request
	conn := NewConn(wsConn)
	// 这里才是 open 事件的入口
	// bind

	if UserID > 10240 {
		UserID = 0
	} else {
		UserID++
	}
	_ = wh.binder.Bind(UserID, conn)
	if wh.onOpen != nil {
		wh.onOpen(conn, UserID)
	}
	conn.AfterReadFunc = func(messageType int, r io.Reader) {
		// 这里是 message 事件
		if wh.onMessage != nil {
			p, err := ioutil.ReadAll(r)
			//_, msg, err := conn.Conn.ReadMessage()
			wh.onMessage(conn, UserID, string(p), err)
		}
	}
	conn.BeforeCloseFunc = func() {
		// unbind 这里是 close  事件
		if wh.onClose != nil {
			wh.onClose(conn, UserID)
		}
		_ = wh.binder.Unbind(conn)
	}

	conn.Listen()
}

// closeConns unbind conns filtered by userID and event and close them.
// The userID can't be empty, but event can be empty. The event will be ignored
// if empty.
func (wh *websocketHandler) closeConns(userID int) (int, error) {
	conns, _ := wh.binder.FilterConn(userID)

	if err := wh.binder.Unbind(conns); err != nil {
		log.Printf("conn unbind fail: %v", err)
		return 0, nil
	}

	return 1, nil
}
