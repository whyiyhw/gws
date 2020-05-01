package main

import (
	"fmt"
	"github.com/whyiyhw/gws"
	"sync"
)

func main() {
	server := new(gws.Server)
	// 需要增加一个 去记录 userID 与 连接的关系
	var mu sync.RWMutex
	relation := map[int]*gws.Conn{} // 这里可以替换成redis

	server.OnMessage = func(conn *gws.Conn, fd int, message string, err error) {

		// 接入后给对应的 连接发 消息
		// response := fmt.Sprintf("had recv you message: %s    : by server default info", message)
		//_, _ = conn.Write([]byte(response))

		fmt.Printf("client %d said %s \n", fd, message)
	}

	server.OnOpen = func(conn *gws.Conn, fd int) {
		fmt.Printf("client %d online \n", fd)

		mu.Lock()
		defer mu.Unlock()
		relation[fd] = conn
		// 获取所有在线的连接
		for k, v := range relation {
			fmt.Println(k)
			_, _ = v.Write([]byte(fmt.Sprintf("client %d online", fd)))
		}

	}

	server.OnClose = func(conn *gws.Conn, fd int) {
		fmt.Printf("client %d had offline \n", fd)
		mu.Lock()
		defer mu.Unlock()
		delete(relation, fd)
		for k, v := range relation {
			fmt.Println(k)
			_, _ = v.Write([]byte(fmt.Sprintf("client %d offline", fd)))
		}
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
