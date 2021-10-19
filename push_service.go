package main

import (
	"bufio"
	"log"
	"net"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

func pushService(conn net.Conn) {
	// 捕捉 painc
	defer func() {
		if err := recover(); err != nil {
			log.Printf("painc-%s", err)
		}
	}()
	defer conn.Close()

	var bytes = make([]byte, 1)

	// 读入文件类型标识长度
	_, err := conn.Read(bytes)
	checkErrorPrint(err)

	bytes = make([]byte, bytes[0])

	// 读入文件类型标识
	conn.Read(bytes)

	// 将文件名转换成字符串形式
	fileType := string(bytes)

	// 将图片字节数组读入 bytes 中
	bytes = make([]byte, 0)
	temp := make([]byte, 4096)

	for n, err := conn.Read(temp); n != 0; {
		checkErrorPrint(err)
		bytes = append(bytes, temp[:n]...)
		n, err = conn.Read(temp)
	}

	// 生成一个文件名（uuid + . + fileType）
	fileName := uuid.New().String() + "." + fileType

	// 将文件名保存到 redis ，便于后续确定文件是否存在
	redisConn := redisPoll.Get()
	defer redisConn.Close()
	if changeRow, err := redis.Int(redisConn.Do("SADD", fileName)); changeRow == 0 || err != nil {
		log.Println(err)
		return
	}

	// 将文件保存到文件夹
	go writeFile(fileName, fileType, bytes)
	// 将文件名写回给客户
	conn.Write([]byte(fileName))
}

func writeFile(fileName string, fileType string, bytes []byte) {
	outPutFile, outPutError := os.OpenFile(*fileRootDir+"/"+fileType+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	checkErrorPrint(outPutError)

	defer outPutFile.Close()

	outPutWrite := bufio.NewWriter(outPutFile)

	_, err := outPutWrite.Write(bytes)
	checkErrorPrint(err)
	outPutWrite.Flush()
}
