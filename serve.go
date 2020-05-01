package gws

import (
	"github.com/gorilla/websocket"

	"net/http"
)

const (
	serverDefaultWSPath = "/ws"
	serverDefaultAddr   = ":9501"
)

var defaultUpgrade = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

//Server 定义运行 websocket 服务 所需的参数
type Server struct {
	// 定义服务监听端口
	Addr string

	// 定义 websocket 服务的路由 , 默认 "/ws".
	WSPath string

	// Upgrader is for upgrade connection to websocket connection using
	// "github.com/gorilla/websocket".
	//
	// If Upgrader is nil, default upgrader will be used. Default upgrader is
	// set ReadBufferSize and WriteBufferSize to 1024, and CheckOrigin always
	// returns true.
	Upgrader *websocket.Upgrader

	wh *websocketHandler

	// 连接事件
	OnOpen func(conn *Conn, fd int)
	// 消息接受事件
	OnMessage func(conn *Conn, fd int, message string, err error)
	// 连接关闭事件
	OnClose func(conn *Conn, fd int)

	Send func(conn *Conn, fd int, message string) (err error)
}

//ListenAndServe 监听tcp 连接并处理  websocket 请求
func (s *Server) ListenAndServe() error {
	b := &binder{
		userID2ConnMap: make(map[int]*Conn),
	}

	// websocket 请求处理结构体
	wh := websocketHandler{
		upgrader: defaultUpgrade,
		binder:   b,
	}
	if s.OnClose != nil {
		wh.onClose = s.OnClose
	}
	if s.OnMessage != nil {
		wh.onMessage = s.OnMessage
	}
	if s.OnOpen != nil {
		wh.onOpen = s.OnOpen
	}

	if s.Upgrader != nil {
		wh.upgrader = s.Upgrader
	}

	if s.WSPath == "" {
		s.WSPath = serverDefaultWSPath
	}
	if s.Addr == "" {
		s.Addr = serverDefaultAddr
	}
	s.wh = &wh
	http.Handle(s.WSPath, s.wh)

	return http.ListenAndServe(s.Addr, nil)
}
