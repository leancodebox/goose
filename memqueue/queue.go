package memqueue

import (
	"errors"
	"sync"

	"github.com/leancodebox/goose/array"
	"github.com/leancodebox/goose/fileopt"
	"github.com/leancodebox/goose/jsonopt"
)

var queueLock = &sync.Mutex{}
var queueSaveLock = &sync.Mutex{}
var queueFilePath string
var queueList = make(map[string][]string)

func InitQueue(path string) {
	queueFilePath = path
	data, _ := fileopt.FileGetContents(queueFilePath)
	fileQueue := jsonopt.Decode[map[string][]string](string(data))
	if fileQueue != nil {
		queueList = fileQueue
	}
}

func saveQueueData() {
	if len(queueFilePath) == 0 {
		return
	}
	queueSaveLock.Lock()
	defer queueSaveLock.Unlock()
	fileopt.PutContent(queueFilePath, jsonopt.EncodeFormat(queueList))
}

func QueueRPush(key string, data ...string) {
	queueLock.Lock()
	defer queueLock.Unlock()
	queue, _ := queueList[key]
	queue = append(queue, data...)
	queueList[key] = queue
	saveQueueData()
}

func QueueLPop(key string) (string, error) {
	queueLock.Lock()
	defer queueLock.Unlock()
	queue, _ := queueList[key]
	if len(queue) > 0 {
		result := queue[0]
		queue = queue[1:]
		queueList[key] = queue
		saveQueueData()
		return result, nil
	}
	return "", errors.New("queue is null")
}

func QueueRPushObj[T any](key string, data ...T) {
	queueLock.Lock()
	defer queueLock.Unlock()
	queue, _ := queueList[key]
	strData := array.ArrayMap(func(t T) string {
		entity := jsonopt.Encode(t)
		return entity
	}, data)
	queue = append(queue, strData...)
	queueList[key] = queue
	saveQueueData()
}

func QueueLPopObj[t any](key string) (t, error) {
	queueLock.Lock()
	defer queueLock.Unlock()
	queue, _ := queueList[key]
	if len(queue) > 0 {
		result := queue[0]
		queue = queue[1:]
		queueList[key] = queue
		saveQueueData()
		return jsonopt.DecodeE[t](result)
	}
	var obj t
	return obj, errors.New("queue is null")
}

func QueueLen(key string) int {
	queueLock.Lock()
	defer queueLock.Unlock()
	queue, ok := queueList[key]
	if ok {
		return len(queue)
	}
	return 0
}
