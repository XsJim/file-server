package main

import (
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	pushServiceListenPort = flag.String("pushPort", ":10375", "推送文件服务监听端口")
	pullServiceListenPort = flag.String("pullPort", ":10376", "拉取文件服务监听端口")

	redisAddr     = flag.String("redisAddr", "59.110.218.108:10324", "redis服务地址")
	redisPassword = flag.String("redisPassword", "S5sdf65FxvhSDFo5", "redis密码")
	// 在程序进行服务前，应该创建这个路径
	//fileRootDir = flag.String("fileRootDir", "/home/xsjim/file_server_dir", "本地文件主目录")
	fileRootDir   = flag.String("fileRootDir", "X:/file", "本地文件主目录")
	emptyTypeFile = flag.String("emptyTypeFile", "emptyType", "用来放置没有类型的文件的文件夹，这个文件夹会被自动创建")

	setName = flag.String("setName", "fileNameSet", "用于存放文件名的集合的名字")
)

var (
	redisPoll *redis.Pool
)

func init() {
	// 打开 redis 连接池
	redisPoll = newPool()
}

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", *redisAddr, redis.DialPassword(*redisPassword))
		},
	}
}

func main() {
	// 打开一个 tcp 监听端口，用于接收客户端代理（我应该实现该代理）发送来的添加文件请求
	pushListen, err := net.Listen("tcp", *pushServiceListenPort)
	checkErrorFatal(err)
	defer pushListen.Close()
	go func() {
		for {
			pushConn, err := pushListen.Accept()
			checkErrorPrint(err)
			go pushService(pushConn)
		}
	}()

	// 之后应该打开一个 http 监听端口，监听拉取请求
	http.HandleFunc("/", PullServiceDoor)
	err = http.ListenAndServe(*pullServiceListenPort, nil)
	checkErrorFatal(err)
}
