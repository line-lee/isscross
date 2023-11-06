package main

import (
	"fmt"
	"sunflower/server/pool"
)

func main() {
	fmt.Println("========================== sunflower start ======================================")
	// 启动连接池心跳循环自检
	pool.TimeWheel()
	// 创建服务，接收连接，填充连接池数据
	pool.Run(24763)

}
