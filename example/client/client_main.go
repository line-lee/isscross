package main

import (
	"log"
	"sunflower/common/models"
	"sunflower/sdk"
)

func main() {
	sdk.New(&models.ClientConfig{
		Network: "tcp",
		Address: "127.0.0.1",
		Port:    24763,
	}).ShareSubscribe(handle)

	<-make(chan bool)

}

func handle(message []byte) {
	log.Printf("客户端收到信息：%s\n", string(message))
}
