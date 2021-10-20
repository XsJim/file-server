package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"io"
	"net/http"
	"os"
	"strings"
)

func PullServiceDoor(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		// 正确的地址应该是 http://127.0.0.1:8080/getFile/uuid.filetype
		url := strings.Split(req.URL.Path[1:], "/")

		// 过滤不正确的请求
		if len(url) != 2 || url[0] != "getFile" || !FileExist(url[1]) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			_, err := fmt.Fprint(w, FileBytes(url[1]))
			checkErrorPrint(err)
		}
	}
}

// FileExist 在 redis 中确认文件是否录入
func FileExist(fileName string) (exist bool) {
	redisConn := redisPoll.Get()
	defer redisConn.Close()
	if re, err := redis.Int(redisConn.Do("SISMEMBER", *setName, fileName)); re == 1 && err == nil {
		exist = true
	}

	return
}

// FileBytes 从文件系统中取出文件 bytes ，这个文件名应该是确定有对应文件存在的
func FileBytes(fileName string) (fileBytes []byte) {
	fileType := fileName[strings.LastIndex(fileName, ".")+1:]
	if fileType == "" {
		fileType = *emptyTypeFile
	}
	file, err := os.Open(*fileRootDir + "/" + fileType + "/" + fileName)
	checkErrorPrint(err)
	defer file.Close()

	fileBytes = make([]byte, 0)
	temp := make([]byte, 2048)

	for n, err := file.Read(temp); err != io.EOF; {
		fileBytes = append(fileBytes, temp[:n]...)
		n, err = file.Read(temp)
	}
	return
}
