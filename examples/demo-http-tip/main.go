package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/whyiyhw/gws"
)

func main() {
	server := new(gws.Server)
	// 需要增加一个 去记录 userID 与 连接的关系
	Relation := map[int]*gws.Conn{} // 这里可以替换成redis

	server.OnMessage = func(conn *gws.Conn, fd int, message string, err error) {

		// 接入后给对应的 连接发 消息
		// response := fmt.Sprintf("had receive you message: %s    : by server default info", message)
		//_, _ = conn.Write([]byte(response))

		fmt.Printf("client %d said %s \n", fd, message)
	}

	server.OnOpen = func(conn *gws.Conn, fd int) {
		fmt.Printf("client %d online \n", fd)
		Relation[fd] = conn
	}

	server.OnClose = func(conn *gws.Conn, fd int) {
		fmt.Printf("client %d had offline \n", fd)
		delete(Relation, fd)
	}

	s := new(gws.HttpHandler)
	// 默认 http 请求 http://127.0.0.1:9501/test 即可给所有客户端发消息(~ v ~)
	s.Path = "/test"
	s.DealFunc = func(w http.ResponseWriter, r *http.Request) {
		// 这里因为没有
		for _, c := range Relation {
			_, _ = c.Write([]byte("this is come from http request info" + time.Now().String()))
		}
		fmt.Printf("http request comein \n")
		_, _ = io.WriteString(w, fmt.Sprintf("success send message to all %d client msg", len(Relation)))
	}

	server.OnHttp = append(server.OnHttp, s)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
