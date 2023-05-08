package luckrand

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandTrue(Molecular int, Denominator int) bool {
	return rand.Intn(Denominator) < Molecular
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var lettersLen = len(letters)

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}
	return string(b)
}

func RandomNum(n int) int {
	return rand.Intn(n)
}

// IdMakerInOnP idMaker
type IdMakerInOnP struct {
	id   uint64
	lock sync.Mutex
	once sync.Once
}

func (itself *IdMakerInOnP) SetStartId(id uint64) {
	itself.once.Do(func() {
		itself.lock.Lock()
		defer itself.lock.Unlock()
		itself.id = id
	})
}

func (itself *IdMakerInOnP) Get() uint64 {
	itself.lock.Lock()
	defer itself.lock.Unlock()
	itself.id += 1
	return itself.id
}

type Counter struct {
	lock   sync.Mutex
	number int64
}

func (itself *Counter) Add() int64 {
	itself.lock.Lock()
	defer itself.lock.Unlock()
	itself.number += 1
	return itself.number
}

func (itself *Counter) Get() int64 {
	itself.lock.Lock()
	defer itself.lock.Unlock()
	return itself.number
}

type Trace struct {
	traceId string
	number  int
}

func (itself *Trace) GetNextTrace() string {
	itself.number += 1
	return fmt.Sprintf("%v:%v", itself.traceId, itself.number)
}

func GetTrace() *Trace {
	return &Trace{
		traceId: uuid.NewString(),
	}
}

func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// 得到id字符串
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

var lock sync.Mutex
var traceManager = map[int]*Trace{}

func MyTraceClean() {
	lock.Lock()
	defer lock.Unlock()
	id := GoID()
	delete(traceManager, id)
}

func MyTrace() *Trace {
	lock.Lock()
	defer lock.Unlock()
	id := GoID()
	if trace, ok := traceManager[id]; ok {
		return trace
	}
	return &Trace{
		traceId: `00000000-0000-0000-0000-000000000000`,
	}
}

func MyTraceInit() *Trace {
	lock.Lock()
	defer lock.Unlock()
	id := GoID()
	traceManager[id] = GetTrace()
	return traceManager[id]
}
