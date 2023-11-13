package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"sunflower/common/models"
	tool "sunflower/common/tools"
)

var thisConnect *models.Client

type sdk struct {
	client *models.Client
}

func New(config *models.ClientConfig) *sdk {
	thisConnect = &models.Client{Config: config}
	thisConnect.Mutex.Lock()
	// 开始监听长连接消息
	tool.SecureGo(func(args ...interface{}) { listen() })
	// 开始心跳监测
	tool.SecureGo(func(args ...interface{}) { heartbeatPush() })
	// 消息重试管理
	tool.SecureGo(func(args ...interface{}) { resend() })
	conn, err := net.Dial(config.Network, fmt.Sprintf("%s:%d", config.Address, config.Port))
	if err != nil {
		log.Printf("sunflower 新建本地链接【%s://%s:%d】错误：%v\n", config.Network, config.Address, config.Port, err)
		thisConnect.Mutex.Unlock()
		reconnect()
		return nil
	}
	thisConnect.Conn = conn

	return &sdk{client: thisConnect}
}

func (s *sdk) ShareSubscribe(f func(message []byte)) {
	tool.SecureGo(func(args ...interface{}) {
		for {
			select {
			case bytes := <-sc:
				f(bytes)
			}
		}
	})
}

func ShareSource(message []byte) {
	if len(message) > 1024 {
		log.Printf("sunflower 应用端发起消息内容过大")
		return
	}
	m := &models.Message{Mid: uuid.NewString(), Types: models.ShareSource, Content: message}
	mem[m.Mid] = m
	bytes, _ := json.Marshal(m)
	write(thisConnect, bytes)
}
