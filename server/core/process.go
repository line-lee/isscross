package core

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"sunflower/common/models"
	"time"
)

func process(cli *models.Client, mb []byte) {
	mm := new(models.Message)
	_ = json.Unmarshal(mb, mm)
	if mm.Types != models.HeartbeatPublish && mm.Types != models.HeartbeatAck {
		log.Printf("收到消息=========>>>>>#%s#\n", string(mb))
	}
	thisTime := time.Now().Unix()
	switch mm.Types {
	case models.HeartbeatPublish:
		// 对方链路心跳问询，需要回执消息，并且更新对方心跳问询时间
		cli.Mutex.Lock()
		defer cli.Mutex.Unlock()
		cli.HeartbeatPullTime = thisTime
		mb, _ := json.Marshal(models.Message{Mid: mm.Mid, Types: models.HeartbeatAck})
		write(cli, models.HeartbeatAck, mb)
	case models.HeartbeatAck:
		// 本方心跳问询，对方回执成功，更新链接对象本方链接问询心跳时间戳
		cli.Mutex.Lock()
		defer cli.Mutex.Unlock()
		cli.HeartbeatPushTime = thisTime
	case models.ShareSource:
		// 收到应用端的发布消息，广播给同组的所有连接
		for _, cp := range clientPool {
			if cp.UUID == cli.UUID {
				continue
			}
			// 复制消息
			push := new(models.Message)
			_ = json.Unmarshal(mb, push)
			push.Mid = uuid.NewString()
			push.Types = models.SharePublish
			// 加入消息重试队列 TODO 要用线程安全的map
			messageMap[push.Mid] = push
			// 发送消息
			pby, _ := json.Marshal(push)
			write(cp, push.Types, pby)
		}
		mm.Types = models.ShareACK
		kb, _ := json.Marshal(mm)
		write(cli, mm.Types, kb)
	case models.ShareACK:
		// 收到消息回执，清理重试map中的key
		delete(messageMap, mm.Mid)
	default:
		log.Printf("消息处理，消息类型解析错误:%v\n", mm.Types)
	}
}
