package main

import (
	"bufio"
	"errors"
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
	_, err = conn.Read(bytes)
	checkErrorPrint(err)
	// 将文件类型标识转换成字符串形式
	fileType := string(bytes)

	// 读取文件 byte 尺寸，长度协商为 4 位
	bytes = make([]byte, 4)
	_, err = conn.Read(bytes)
	checkErrorPrint(err)
	byteSize := ByteSizeFromByteSlice(bytes)

	// 将图片字节数组读入 bytes 中
	bytes = make([]byte, byteSize)
	_, err = conn.Read(bytes)
	checkErrorPrint(err)
	// 生成一个文件名（uuid + . + fileType）
	fileName := uuid.New().String() + "." + fileType

	// 将文件名保存到 redis ，便于后续确定文件是否存在
	redisConn := redisPoll.Get()
	defer redisConn.Close()
	if changeRow, err := redis.Int(redisConn.Do("SADD", *setName, fileName)); changeRow == 0 || err != nil {
		log.Println(err)
		return
	}

	// 将文件保存到文件夹
	go writeFile(fileName, fileType, bytes)
	// 将文件名写回给客户，加上文件类型
	_, err = conn.Write([]byte(fileName))
	checkErrorPrint(err)
}

func writeFile(fileName string, fileType string, bytes []byte) {
	if fileType == "" {
		fileType = *emptyTypeFile
	}
	// 首先检查是否有该类型的文件夹，如果没有，创建一个
	if !DirExist(fileType) {
		MkDir(fileType)
	}
	outPutFile, outPutError := os.OpenFile(*fileRootDir+"/"+fileType+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if outPutError != nil {
		log.Println(outPutError)
		return
	}

	defer outPutFile.Close()

	outPutWrite := bufio.NewWriter(outPutFile)

	_, err := outPutWrite.Write(bytes)
	checkErrorPrint(err)
	err = outPutWrite.Flush()
	checkErrorPrint(err)
}

// DirExist 检查文件夹是否存在
func DirExist(dirName string) (exist bool) {
	_, err := os.Stat(*fileRootDir + "/" + dirName)
	return err == nil || errors.Is(err, os.ErrExist)
}

// MkDir 创建一个不存在的目录
func MkDir(dirName string) {
	err := os.Mkdir(*fileRootDir+"/"+dirName, 0666)
	checkErrorPrint(err)
}

// ByteSizeFromByteSlice 从一个 byte 切片中组合出一个表示字节数量的 int
func ByteSizeFromByteSlice(tar []byte) (byteSize int) {
	for i := 0; i < 4; i++ {
		byteSize |= int(tar[i]) << (8 * i)
	}

	return
}
