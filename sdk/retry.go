package sdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sunflower/common/models"
	"time"
)

var retry = make(chan int)

func reconnect() {
	log.Printf("sunflower 触发重连机制================>>>>>>>>\n")
	for {
		thisConnect.Mutex.Lock()
		conn, err := net.DialTimeout(thisConnect.Config.Network, fmt.Sprintf("%s:%d", thisConnect.Config.Address, thisConnect.Config.Port), 3*time.Second)
		if err != nil {
			log.Printf("sunflower 触发重连机制，新建链接报错：%v", err)
			thisConnect.Mutex.Unlock()
			time.Sleep(time.Second)
			continue
		}
		thisConnect.Conn = conn
		var thisTime = time.Now().Unix()
		thisConnect.HeartbeatPullTime = thisTime
		thisConnect.HeartbeatPushTime = thisTime
		thisConnect.Mutex.Unlock()
		log.Printf("sunflower 重连成功！！！！！！！！！！！！！！\n")
		return
	}
}

var mem = make(map[string]*models.Message)

func resend() {
	for key, message := range mem {
		message.Mutex.Lock()
		const retryLimit = 10 // 重试阈值
		if message.Retry >= retryLimit {
			log.Printf("suflower 重试消息【%s】，重试次数【%d】，删除消息内容：%v", key, message.Retry, message)
			delete(mem, key)
			message.Mutex.Unlock()
			continue
		}
		bytes, _ := json.Marshal(message)
		write(thisConnect, bytes)
		message.Mutex.Unlock()
	}
}

func write(client *models.Client, bytes []byte) {
	log.Printf("suflower 发送消息=========>>>>>#%s#", string(bytes))
	_, err := client.Conn.Write(bytes)
	if err != nil {
		log.Printf("suflower 向服务端发送消息错误：%v\n", err)
		reconnect()
		return
	}
}
