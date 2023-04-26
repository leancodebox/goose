package power

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// GuardGoRoutine 守护并恢复异常中断的
func GuardGoRoutine(action func()) {
	go func() {
		for {
			restart := upAction(action)
			if restart == false {
				break
			}
		}
	}()
}

func upAction(action func()) (restart bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			restart = true
		}
	}()
	action()
	restart = false
	return
}

// Together 并行执行
func Together(job func(goId int), counter int) {
	var wg sync.WaitGroup
	for i := 1; i <= counter; i++ {
		wg.Add(1)
		go func(i int) {
			minDuration := time.Duration(1) // 最小等待时间为 1 秒
			maxDuration := time.Duration(2) // 最大等待时间为 2 秒
			waitingTime := minDuration + time.Duration(rand.Int63n(int64(maxDuration-minDuration)))*time.Second
			time.Sleep(waitingTime)
			defer wg.Done()
			job(i)
		}(i)
	}
	wg.Wait()
}

// 如果之后非常大要在锁内加一个限制，
// 超过一定量就一律返回true,
// 同时删除逻辑中暂不使用重建map等回收 map 的方式，
// 这些方式只可以在突破内存预定峰值后进行一个内存的回落处理，
// 但是并不能作到内存使用峰值的降低。后续如果有这类场景，可以考虑持久化或更换为redis，
// 不必过多对forbidden代码进行优化。
var forbiddenList = make(map[string]int64, 256)
var forbiddenLock sync.Mutex

func init() {
	// 10s 执行一次，清理过期的key
	go func() {
		for {
			time.Sleep(time.Second * 10)
			clearForbiddenList()
		}
	}()
}

// 清理过期的key,不会释放已经申请的内存，但是可以使已经过期的内存在之后重新被利用
func clearForbiddenList() {
	forbiddenLock.Lock()
	defer forbiddenLock.Unlock()
	nowTS := time.Now().UnixMilli()
	for key, data := range forbiddenList {
		if data < nowTS {
			delete(forbiddenList, key)
		}
	}
}

// Forbidden 非单一部署情形下，不要使用该函数
func Forbidden(key string, timeout int) bool {
	forbiddenLock.Lock()
	defer forbiddenLock.Unlock()
	if _, ok := forbiddenList[key]; ok {
		return false
	} else {
		forbiddenList[key] = time.Now().Add(time.Duration(timeout) * time.Second).UnixMilli()
		return true
	}
}

func AllTrue(conditionList ...bool) bool {
	for _, cond := range conditionList {
		if cond == false {
			return false
		}
	}
	return true
}

func HaveTrue(conditionList ...bool) bool {
	for _, cond := range conditionList {
		if cond == true {
			return true
		}
	}
	return false
}

func Eq[T comparable](a, b T) bool {
	return a == b
}
