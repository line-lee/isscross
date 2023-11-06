package pool

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	tool "sunflower/common/tools"
	"sync"
	"time"
)

var clientPool = make(map[string]*client)

type client struct {
	// tcp连接
	conn net.Conn
	// 链接唯一标识
	uuid string
	// 互斥锁
	mutex sync.Mutex
	// 链接对方问询心跳时间戳
	heartbeatPullTime int64
	// 链接本方问询心跳，对方回应时间戳
	heartbeatPushTime int64
	// 使用context.WithCancel()方法创建一个带有取消信号的上下问对象
	// 上下文对象， 连接中使用ctx.Done监听到连接失效，中断循环读取连接传输内容
	ctx context.Context
	// 上下文触发取消信号函数
	cancel context.CancelFunc
}

func TimeWheel() {
	tool.SecureGo(func(args ...interface{}) {
		for {
			// 心跳监测结果处理
			tool.SecureGo(func(args ...interface{}) {
				for k, v := range clientPool {
					tool.SecureGo(func(args ...interface{}) {
						key := args[0].(string)
						cli := args[1].(*client)
						heartbeatCheck(key, cli)
					}, k, v)
				}
			})
			// 消息重试结果处理
			tool.SecureGo(func(args ...interface{}) {
				for k, v := range messageMap {
					tool.SecureGo(func(args ...interface{}) {
						key := args[0].(string)
						msg := args[1].(*message)
						retry(key, msg)
					}, k, v)
				}
			})
		}

	})
	// 时间轮一分钟检查一次
	time.Sleep(time.Minute)
}

func heartbeatCheck(key string, cli *client) {
	cli.mutex.Lock()
	defer cli.mutex.Unlock()
	thisTime := time.Now().Unix()
	const heartbeatTimeLimit = 60 // 心跳监测断链阈值
	if thisTime-cli.heartbeatPullTime > heartbeatTimeLimit {
		log.Fatalf("tcp链接【%s】对方超过【%d】秒没有主动问询，判定断开连接\n", key, thisTime-cli.heartbeatPullTime)
		cli.cancel()
		return
	}
	if thisTime-cli.heartbeatPushTime > heartbeatTimeLimit {
		log.Fatalf("tcp链接【%s】本方问询，超过【%d】秒没有回复，判定断开连接\n", key, thisTime-cli.heartbeatPushTime)
		cli.cancel()
		return
	}
}

func retry(key string, msg *message) {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()
	const retryLimit = 10 // 重试阈值
	if msg.retry >= retryLimit {
		log.Fatalf("重试消息【%s】，重试次数【%d】，删除消息内容：%v", key, msg.retry, msg)
		delete(messageMap, key)
		return
	}
}

func Run(port int) {
	fmt.Println("==========================sunflower start======================================")
	listen, err := net.Listen("tcp", ":24763")
	if err != nil {
		log.Fatalf("项目启动tcp端口监听报错：%v\n", err)
		return
	}
	defer listen.Close()
	//循环等待客户端来连接
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalf("监听tcp连接错误:%v\n", err)
			continue
		}
		// 1.加入连接池，让定时器循环检测
		ctx, cancel := context.WithCancel(context.Background())
		cli := &client{conn: conn, uuid: uuid.NewString(), ctx: ctx, cancel: cancel}
		clientPool[cli.uuid] = cli
		// 开启连接消息监听协程
		tool.SecureGo(func(args ...interface{}) {
			nc := args[0].(*client)
			cc := args[1].(context.Context)
			subscribe(nc, cc)
		}, cli, ctx)
	}
}

func subscribe(cli *client, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			//上下文对象接收到关闭信号.
			//证明在连接池clientPool管理中，发现client对象的心跳连接出现问题，触发cancel函数.
			//此时我们认为这个连接已经不具备活性，应当关闭
			log.Fatalln("tcp连接订阅，服务器心跳监测失去活性，断开连接")
			cli.conn.Close()
			return
		default:
			// 使用1024字节缓冲流对象接收监听消息
			var buf = make([]byte, 1024)
			_, err := cli.conn.Read(buf)
			if err != nil {
				log.Fatalln("tcp连接订阅，读取消息内容错误", err)
				continue
			}
			tool.SecureGo(func(args ...interface{}) {
				// 协程处理消息内容
				c := args[0].(*client)
				bytes := args[1].([]byte)
				process(c, bytes)
			}, cli, buf)

		}
	}
}

type MessageType int

var (
	HeartbeatPublish MessageType = 1 // 发起心跳问询消息
	HeartbeatAck     MessageType = 2 // 回执心跳问询消息

	ShareSource  MessageType = 10 // 共享消息来源
	SharePublish MessageType = 11 // 共享消息发布
	ShareACK     MessageType = 12 // 共享消息回执
)

var messageMap = make(map[string]*message)

type message struct {
	topic string      // 相同topic服务才共享信息 TODO
	mid   string      // 消息唯一标识，用于监测消息发送，到达，重试等操作
	types MessageType // 消息类型：发布，回执
	msg   []byte      // 消息内容
	retry int         // 重试次数
	mutex sync.Mutex  // 互斥锁
}

func process(cli *client, bytes []byte) {
	m := new(message)
	json.Unmarshal(bytes, m)
	thisTime := time.Now().Unix()
	switch m.types {
	case HeartbeatPublish:
		// 对方链路心跳问询，需要回执消息，并且更新对方心跳问询时间
		cli.mutex.Lock()
		defer cli.mutex.Unlock()
		cli.heartbeatPullTime = thisTime
		mb, _ := json.Marshal(message{mid: m.mid, types: HeartbeatAck})
		cli.conn.Write(mb)
	case HeartbeatAck:
		// 本方心跳问询，对方回执成功，更新链接对象本方链接问询心跳时间戳
		cli.mutex.Lock()
		defer cli.mutex.Unlock()
		cli.heartbeatPushTime = thisTime
	case ShareSource:
		// 收到应用端的发布消息，广播给同组的所有连接
		for _, cp := range clientPool {
			cp.mutex.Lock()
			// 复制消息
			push := new(message)
			json.Unmarshal(bytes, push)
			push.mid = uuid.NewString()
			push.types = SharePublish
			// 加入消息重试队列
			messageMap[push.mid] = push
			// 发送消息
			pby, _ := json.Marshal(push)
			cp.conn.Write(pby)
			cp.mutex.Unlock()
		}
	case ShareACK:
		// 收到消息回执，清理重试map中的key
		delete(messageMap, m.mid)
	default:
		log.Fatalf("消息处理，消息类型解析错误:%v/\n", m.types)
	}
}
