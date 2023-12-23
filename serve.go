package gws

import (
	"net/http"

	"github.com/gorilla/websocket"
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

// Server 定义运行 websocket 服务 所需的参数
type Server struct {
	// 定义服务监听端口
	Addr string

	// 定义 websocket 服务的路由 , 默认 "/ws".
	WSPath string

	// websocket 的升级主要通过 以下 包来实现
	// "github.com/gorilla/websocket".
	//
	// 默认的 upgrader  ReadBufferSize 和 WriteBufferSize 都是 1024  CheckOrigin 方法 为不检查
	Upgrader *websocket.Upgrader

	wh *websocketHandler

	// 连接事件
	OnOpen func(conn *Conn, fd int)
	// 消息接受事件
	OnMessage func(conn *Conn, fd int, message string, err error)
	// 连接关闭事件
	OnClose func(conn *Conn, fd int)

	OnHttp []*HttpHandler
}

// ListenAndServe 监听tcp 连接并处理  websocket 请求
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

	// 新增 http 请求的处理
	if len(s.OnHttp) > 0 {
		for _, v := range s.OnHttp {
			http.Handle(v.Path, v.DealFunc)
		}
	}

	return http.ListenAndServe(s.Addr, nil)
}
