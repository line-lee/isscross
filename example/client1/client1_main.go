package main

import (
	"context"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "10.10.1.117:24763")
	if err != nil {
		log.Fatalln("客户端创建tcp连接错误")
		return
	}
	defer conn.Close()

	// 发消息
	conn.Write([]byte("123"))
	context.WithCancel()

}
