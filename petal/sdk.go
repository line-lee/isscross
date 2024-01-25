package petal

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/line-lee/isscross/common/models"
	tool "github.com/line-lee/isscross/common/tools"
	"log"
)

var thisConnect *models.Client

type petal struct {
	client *models.Client
}

func Start(config *models.ClientConfig) *petal {
	thisConnect = &models.Client{Config: config}
	// 开始监听长连接消息
	tool.SecureGo(func(args ...interface{}) { listen() })
	//开始心跳监测
	tool.SecureGo(func(args ...interface{}) { heartbeatPush() })
	// 消息重试管理
	tool.SecureGo(func(args ...interface{}) { resend() })
	return &petal{client: thisConnect}
}

func (p *petal) Listen(handle func(message []byte)) {
	tool.SecureGo(func(args ...interface{}) {
		for {
			select {
			case bytes := <-sc:
				handle(bytes)
			}
		}
	})
}

func Share(message []byte) {
	if len(message) > 1024 {
		log.Printf("isscross 应用端发起消息内容过大\n")
		return
	}
	m := &models.Message{Mid: uuid.NewString(), Types: models.ShareSource, Content: message}
	mem[m.Mid] = m
	bytes, _ := json.Marshal(m)
	write(thisConnect, m.Types, bytes)
}
