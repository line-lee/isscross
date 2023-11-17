package petal

import (
	"encoding/json"
	"log"
	"regexp"
	"sunflower/common/models"
	"time"
)

func listen() {
	for {
		if thisConnect.Conn == nil {
			reconnect()
			time.Sleep(time.Second)
			continue
		}
		var buf = make([]byte, 1024)
		_, err := thisConnect.Conn.Read(buf)
		if err != nil {
			log.Printf("sunflower 应用端使用长连接读取数据报错，链接中断：%v", err)
			reconnect()
			continue
		}
		compile, _ := regexp.Compile(`{.*}`)
		str := compile.FindString(string(buf))
		process([]byte(str))
	}
}

// 订阅服务端推送的广播消息
var sc = make(chan []byte, 102400)

func process(buf []byte) {
	mm := new(models.Message)
	_ = json.Unmarshal(buf, mm)
	if mm.Types != models.HeartbeatPublish && mm.Types != models.HeartbeatAck {
		log.Printf("sunflower 收到消息=========>>>>>#%s#\n", string(buf))
	}
	thisTime := time.Now().Unix()
	switch mm.Types {
	case models.HeartbeatPublish:
		// 对方链路心跳问询，需要回执消息，并且更新对方心跳问询时间
		thisConnect.Mutex.Lock()
		defer thisConnect.Mutex.Unlock()
		thisConnect.HeartbeatPullTime = thisTime
		mm.Types = models.HeartbeatAck
		bytes, _ := json.Marshal(mm)
		write(thisConnect, mm.Types, bytes)
	case models.HeartbeatAck:
		// 本方心跳问询，对方回执成功，更新链接对象本方链接问询心跳时间戳
		thisConnect.Mutex.Lock()
		defer thisConnect.Mutex.Unlock()
		thisConnect.HeartbeatPushTime = thisTime
	case models.SharePublish:
		// 服务端发起内存共享，收到执行，并回复确认
		sc <- mm.Content
		mm.Types = models.ShareACK
		bytes, _ := json.Marshal(mm)
		write(thisConnect, mm.Types, bytes)
	case models.ShareACK:
		// 收到内存共享信息回执，删除消息重试map
		delete(mem, mm.Mid)
	default:
		log.Printf("sunflower 应用端解析消息类型错误：%v\n", mm.Types)
		return
	}
}
