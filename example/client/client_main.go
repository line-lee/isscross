package main

import (
	"log"
	"sunflower/common/models"
	"sunflower/petal"
)

func main() {
	petal.Start(&models.ClientConfig{Network: "tcp", Address: "127.0.0.1", Port: 24763}).Listen(handle)

	<-make(chan bool)
}

func handle(message []byte) {
	log.Printf("客户端收到信息：%s\n", string(message))
}
