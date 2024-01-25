package main

import (
	"fmt"
	"github.com/line-lee/isscross/common/models"
	tool "github.com/line-lee/isscross/common/tools"
	"github.com/line-lee/isscross/petal"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestClient1(t *testing.T) {
	rand.NewSource(time.Now().UnixNano())
	petal.Start(&models.ClientConfig{Address: "127.0.0.1:24763"}).Listen(client1Handle)
	tool.SecureGo(func(args ...interface{}) {
		client1Share()
	})
	<-make(chan bool)
}

func client1Share() {
	for {
		time.Sleep(time.Second)
		num := rand.Int63()
		if num%10 == 1 {
			// 当个位数是7的时候触发广播
			str := fmt.Sprintf("Client1，分享信息=====>它找到随机数[%d]", num)
			petal.Share([]byte(str))
		}
	}
}

func client1Handle(message []byte) {
	log.Printf("Client1，监听到消息[%s]\n", string(message))
}

func TestClient2(t *testing.T) {
	rand.NewSource(time.Now().UnixNano())
	petal.Start(&models.ClientConfig{Address: "127.0.0.1:24763"}).Listen(client2Handle)
	tool.SecureGo(func(args ...interface{}) {
		client2Share()
	})
	<-make(chan bool)
}

func client2Share() {
	for {
		time.Sleep(time.Second)
		num := rand.Int63()
		if num%10 == 2 {
			// 当个位数是7的时候触发广播
			str := fmt.Sprintf("Client2，分享信息=====>它找到随机数[%d]", num)
			petal.Share([]byte(str))
		}
	}
}

func client2Handle(message []byte) {
	log.Printf("Client2，监听到消息[%s]\n", string(message))
}
