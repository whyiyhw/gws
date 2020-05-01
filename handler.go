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
		userID, _ := wh.binder.FindIDByConn(conn)
		wh.onOpen(conn, userID)
	}
	conn.AfterReadFunc = func(messageType int, r io.Reader) {
		// 这里是 message 事件
		if wh.onMessage != nil {
			p, err := ioutil.ReadAll(r)
			//_, msg, err := conn.Conn.ReadMessage()
			userID, _ := wh.binder.FindIDByConn(conn)
			wh.onMessage(conn, userID, string(p), err)
		}
	}
	conn.BeforeCloseFunc = func() {
		// unbind 这里是 close  事件
		if wh.onClose != nil {
			// 此时的 userID 不是当时的  userID 了
			// 需要根据 conn 获取 userID
			userID, _ := wh.binder.FindIDByConn(conn)
			wh.onClose(conn, userID)
		}
		_ = wh.binder.Unbind(conn)
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

// pushHandler defines to handle push message request.
type HttpHandler struct {
	Path string
	//binder *binder 内部不提供 对于 内部关系的外部接口
	DealFunc http.HandlerFunc
}

//func (h *HttpHandler) GetBinder() *binder {
//	return h.binder
//}
