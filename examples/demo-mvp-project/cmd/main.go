package main

import (
	"github.com/whyiyhw/gws/examples/demo-mvp-project/cmd/response"
	"math/rand"

	"github.com/whyiyhw/gws"
	"github.com/whyiyhw/gws/examples/demo-mvp-project/cmd/bucket"
	"github.com/whyiyhw/gws/examples/demo-mvp-project/cmd/crontab"
	"github.com/whyiyhw/gws/examples/demo-mvp-project/cmd/routers"
)

func main() {

	server := new(gws.Server)

	// 未认证的bucket 默认初始化10个空闲map
	unAuthBuckets := bucket.NewMapBucket(10)

	// 已认证的bucket 默认初始化10个空闲map
	authBuckets := bucket.NewMapBucket(10)

	// 设置 auth 函数
	authRoute := routers.AuthFunc{
		GetIdByToken: func(token string) (int64, error) {
			// todo 实际上 这里应该是  rpc/api/db 中获取并验证用户数据
			return int64(rand.Intn(1000000000)), nil
		},
	}

	// 接收消息事件
	server.OnMessage = func(conn *gws.Conn, fd int, message string, err error) {

		// 未认证的 bucket 里面
		if unAuthBuckets.Exist(fd) {
			authErr := routers.ParseAndAuth(message, fd, conn, unAuthBuckets, authBuckets, authRoute) // 接收后，解析对应连接发过来的消息
			if authErr != nil {
				_, _ = conn.Write(response.Error(501, authErr.Error()))
				return
			}
			_, _ = conn.Write(response.Success(201))
			return
		}

		// 已认证的 bucket 里面
		if authBuckets.Exist(fd) {
			heartbeatErr := routers.ParseAndExec(message, fd, authBuckets) // 接收后，解析对应连接发过来的消息
			if heartbeatErr != nil {
				_, _ = conn.Write(response.Error(502, heartbeatErr.Error()))
				return
			}
			_, _ = conn.Write(response.Success(202))
			return
		}

		// 其它情况,不在维护范围内，不处理
	}

	// 连接事件
	server.OnOpen = func(conn *gws.Conn, fd int) {
		unAuthBuckets.Set(fd, bucket.New(conn, fd)) // 连接后，把 fd 存入未授权的 buckets
	}

	// 关闭事件
	server.OnClose = func(conn *gws.Conn, fd int) {
		// 解除 授权/未授权 bucket 关联
		unAuthBuckets.Delete(fd)
		authBuckets.Delete(fd)
	}

	// http处理事件
	server.OnHttp = append(server.OnHttp, routers.GetHttpHandleList(authBuckets, unAuthBuckets)...)

	// 定时器-清理 未授权/未响应 连接
	go crontab.TimingClearConn(unAuthBuckets, authBuckets)

	// 启动服务
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
