package main

import (
	"fmt"
	"github.com/line-lee/isscross/server/core"
)

func main() {
	fmt.Println("========================== isscross start ======================================")
	// 启动连接池心跳循环自检
	core.TimeWheel()
	// 创建服务，接收连接，填充连接池数据

	// 这里服务的运行端口默认是24763，可以做修改
	core.Run(24763)
}
