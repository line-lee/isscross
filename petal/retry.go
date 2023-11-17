package petal

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sunflower/common/models"
	"sync"
	"time"
)

var retry = make(chan int)

var connectLock sync.Mutex

func reconnect() {
	log.Printf("sunflower 触发重连机制================>>>>>>>>\n")
	for {
		connectLock.Lock()
		var thisTime = time.Now().Unix()
		const reconnectLimit = 60 // 重连成功后，再次重连得等1min
		if thisTime-thisConnect.ReconnectTime < reconnectLimit {
			log.Printf("sunflower 触发重连机制，重连成功[%d]秒，不再次触发\n", thisTime-thisConnect.ReconnectTime)
			return
		}
		conn, err := net.DialTimeout(thisConnect.Config.Network, fmt.Sprintf("%s:%d", thisConnect.Config.Address, thisConnect.Config.Port), 3*time.Second)
		if err != nil {
			log.Printf("sunflower 触发重连机制，新建链接报错：%v\n", err)
			connectLock.Unlock()
			time.Sleep(time.Second)
			continue
		}
		thisConnect.Conn = conn
		thisConnect.HeartbeatPullTime = thisTime
		thisConnect.HeartbeatPushTime = thisTime
		thisConnect.ReconnectTime = thisTime
		connectLock.Unlock()
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
		mem[key].Retry++
		bytes, _ := json.Marshal(message)
		write(thisConnect, message.Types, bytes)
		message.Mutex.Unlock()
	}
}

func write(client *models.Client, mt models.MessageType, bytes []byte) {
	if mt != models.HeartbeatPublish && mt != models.HeartbeatAck {
		// 打印心跳消息
		log.Printf("suflower 发送消息=========>>>>>#%s#", string(bytes))
	}
	if client.Conn == nil {
		log.Printf("suflower 向服务端发送消息，连接为空\n")
		reconnect()
		return
	}
	_, err := client.Conn.Write(bytes)
	if err != nil {
		log.Printf("suflower 向服务端发送消息错误：%v\n", err)
		reconnect()
		return
	}
}
