package core

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"regexp"
	"sunflower/common/models"
	tool "sunflower/common/tools"
	"time"
)

// 监测所有客户端连接的连接池
var clientPool = make(map[string]*models.Client)

// 监测所有消息的发送、接收以及重试
var messageMap = make(map[string]*models.Message)

func Run(port int) {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("项目启动tcp端口监听报错：%v\n", err)
		return
	}
	defer listen.Close()
	//循环等待客户端来连接
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("监听tcp连接错误:%v\n", err)
			continue
		}
		thisTime := time.Now().Unix()
		cli := &models.Client{Conn: conn, UUID: uuid.NewString(), HeartbeatPushTime: thisTime, HeartbeatPullTime: thisTime}
		clientPool[cli.UUID] = cli
		// 开启连接消息监听协程
		tool.SecureGo(func(args ...interface{}) {
			nc := args[0].(*models.Client)
			subscribe(nc)
		}, cli)
	}
}

func subscribe(cli *models.Client) {
	defer cli.Conn.Close()
	for {
		var buf = make([]byte, 1024)
		_, err := cli.Conn.Read(buf)
		if err != nil {
			//ATTENTION：这里链接一旦被断开（心跳监测没有通过校验），就会报错。中断整个长连接的处理方法
			log.Printf("tcp连接订阅，读取消息内容错误:%v", err)
			return
		}
		compile, _ := regexp.Compile(`{.*}`)
		str := compile.FindString(string(buf))
		process(cli, []byte(str))
	}
}
