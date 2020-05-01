package main

import (
	"fmt"
	"github.com/whyiyhw/gws"
)

func main() {
	server := new(gws.Server)

	server.OnMessage = func(conn *gws.Conn, fd int, message string, err error) {

		// 接入后给对应的 连接发 消息
		// response := fmt.Sprintf("had recv you message: %s    : by server default info", message)
		//_, _ = conn.Write([]byte(response))

		fmt.Printf("client %d said %s \n", fd, message)
	}

	server.OnOpen = func(conn *gws.Conn, fd int) {
		fmt.Printf("client %d online \n", fd)
	}

	server.OnClose = func(conn *gws.Conn, fd int) {
		fmt.Printf("client %d had offline \n", fd)
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
