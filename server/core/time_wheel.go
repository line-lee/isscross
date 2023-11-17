package core

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"sunflower/common/models"
	tool "sunflower/common/tools"
	"time"
)

func TimeWheel() {
	tool.SecureGo(func(args ...interface{}) {
		for {
			// 时间轮一分钟检查一次
			time.Sleep(time.Minute)

			// 心跳监测结果处理
			tool.SecureGo(func(args ...interface{}) {
				for k, v := range clientPool {
					tool.SecureGo(func(args ...interface{}) {
						key := args[0].(string)
						cli := args[1].(*models.Client)
						heartbeatCheck(key, cli)
					}, k, v)
				}
			})
			// 消息重试结果处理
			tool.SecureGo(func(args ...interface{}) {
				for k, v := range messageMap {
					tool.SecureGo(func(args ...interface{}) {
						key := args[0].(string)
						msg := args[1].(*models.Message)
						retry(key, msg)
					}, k, v)
				}
			})
		}
	})
	tool.SecureGo(func(args ...interface{}) {
		for {
			// 时间轮隔20秒发起心跳检查
			time.Sleep(20 * time.Second)
			for _, client := range clientPool {
				if client == nil {
					log.Printf("服务端全检链接心跳，发现空连接")
					continue
				}
				// 心跳监测
				bytes, _ := json.Marshal(models.Message{Mid: uuid.NewString(), Types: models.HeartbeatPublish})
				write(client, models.HeartbeatPublish, bytes)
				// 信息重试
				for _, message := range messageMap {
					mb, _ := json.Marshal(message)
					write(client, message.Types, mb)
				}
			}
		}
	})
}

func heartbeatCheck(key string, cli *models.Client) {
	thisTime := time.Now().Unix()
	const heartbeatTimeLimit = 60 // 心跳监测断链阈值
	if thisTime-cli.HeartbeatPullTime > heartbeatTimeLimit {
		log.Printf("tcp链接【%s】对方超过【%d】秒没有主动问询，判定断开连接\n", key, thisTime-cli.HeartbeatPullTime)
		_ = cli.Conn.Close()
		return
	}
	if thisTime-cli.HeartbeatPushTime > heartbeatTimeLimit {
		log.Printf("tcp链接【%s】本方问询，超过【%d】秒没有回复，判定断开连接\n", key, thisTime-cli.HeartbeatPushTime)
		_ = cli.Conn.Close()
		return
	}
}

func retry(key string, msg *models.Message) {
	msg.Mutex.Lock()
	defer msg.Mutex.Unlock()
	const retryLimit = 10 // 重试阈值
	if msg.Retry >= retryLimit {
		log.Printf("重试消息【%s】，重试次数【%d】，删除消息内容：%v", key, msg.Retry, msg)
		delete(messageMap, key)
		return
	}
}

func write(client *models.Client, mt models.MessageType, bytes []byte) {
	if mt != models.HeartbeatPublish && mt != models.HeartbeatAck {
		// 打印心跳消息
		log.Printf("发送消息=========>>>>>#%s#", string(bytes))
	}
	_, err := client.Conn.Write(bytes)
	if err != nil {
		// 删除链接
		log.Printf("发送消息，错误：%v\n", err)
		delete(clientPool, client.UUID)
	}
}
