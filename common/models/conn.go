package models

import (
	"net"
	"sync"
)

type Client struct {
	// 链接配置信息
	Config *ClientConfig
	// tcp连接
	Conn net.Conn
	// 连接成功
	IsConnectSuccess bool
	// 链接唯一标识
	UUID string
	// 互斥锁
	Mutex sync.Mutex
	// 链接对方问询心跳时间戳
	HeartbeatPullTime int64
	// 链接本方问询心跳，对方回应时间戳
	HeartbeatPushTime int64
}

type ClientConfig struct {
	Network string
	Address string
	Port    int
}

type MessageType string

var (
	HeartbeatPublish MessageType = "HeartbeatPublish" // 发起心跳问询消息
	HeartbeatAck     MessageType = "HeartbeatAck"     // 回执心跳问询消息

	ShareSource  MessageType = "ShareSource"  // 共享消息来源
	SharePublish MessageType = "SharePublish" // 共享消息发布
	ShareACK     MessageType = "ShareACK"     // 共享消息回执
)

type Message struct {
	Topic   string      `json:"topic,omitempty"`   // 相同topic服务才共享信息 TODO
	Mid     string      `json:"mid"`               // 消息唯一标识，用于监测消息发送，到达，重试等操作
	Types   MessageType `json:"types"`             // 消息类型：发布，回执
	Content []byte      `json:"content,omitempty"` // 消息内容
	Retry   int         `json:"retry,omitempty"`   // 重试次数
	Mutex   sync.Mutex  `json:"mutex,omitempty"`   // 互斥锁
}
