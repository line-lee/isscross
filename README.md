# isscross
使用golang开发的，一个轻量级消息广播项目。服务和客户端双向链接，形成太阳花的形状，做到每个花瓣信息共享

# 快速使用
##服务端部署
```
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

go version go1.18.5 linux/amd64
go mod tidy
go build -i -o ./bin/isscross  server/main.go
```
##客户端调用
```
package main

import (
	"github.com/line-lee/isscross/common/models"
	"github.com/line-lee/isscross/petal"
)

func main() {
	// 配置部署好的服务端地址，一定记得要开tcp端口，例如：127.0.0.1:24763
	petal.Start(&models.ClientConfig{Address: "your address"}).Listen(handle)

	// ..............

	// 客户端向其他连接分享消息示例
	str := "share anything"
	petal.Share([]byte(str))

	// run main   .......
}

func handle(message []byte) {
	// listen message to do something...........

}
```

