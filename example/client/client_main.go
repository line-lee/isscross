package main

import (
	"sunflower/common/models"
	"sunflower/petal"
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
