package filequeue

import (
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// NewDefaultFileQueue 标准队列实体，返回一个可以使用的队列管理器
func NewDefaultFileQueue(dirPath string) (*FileQueue, error) {
	return NewFileQueue(dirPath, defaultBlockLen)
}

// NewFileQueue dirPath 是队列目录，blockLen 是队列的块长度，不要小于等于9。前9位用于块数据是否消费/块有效长度等数据。
func NewFileQueue(dirPath string, blockLen int64) (*FileQueue, error) {
	if blockLen < 10 {
		return nil, errors.New("blockLen must be greater than 9 because the first 9 bytes are used to store database block data")
	}
	tmp := FileQueue{queueDir: dirPath,
		header: &QueueHeader{
			version:    libVersion,
			blockLen:   blockLen,
			offset:     defaultOffset,
			dataMaxLen: blockLen - blockValidLen - blockDataLenConfigOffLen, // blockLen - validLen - dateLenConfigLen
		}}
	err := tmp.init()
	return &tmp, err
}

/*
*
head
这里所用的都是字节(byte) 非位(bit)
|(64B) :libVersion(8B) blockLen(8B) offset(8B) 0(8B) 0(8B) 0(8B) 0(8B) 0(8B) |
head libVersion 为版本 blockLen 为块大小 决定后续每个数据块的大小 offset 为当前偏移量，表示着当前位于队列的哪一个数据块下
如果为0 则说明位于第一个数据块下 为 headLen + 0 * blockLen = 64
|(64B): valid(1B) len(8B) data(小于55B) 0(xB)|
数据块格式 第一位为预设有效位。第2到9字节为当前数据长度。表示从 headLen + 0 * blockLen + 1 + 8 开始 取 len 长度的字节为之前存储的数据
|(64B): valid(1B) len(8B) data(小于55B) 0(xB)|
|(64B): valid(1B) len(8B) data(小于55B) 0(xB)|
|(64B): valid(1B) len(8B) data(小于55B) 0(xB)|
*/

// 下面部分常量不会记录在文件中
const (
	// headLen head 长度 文件前 xB 的数据为header 的存储空间
	headLen int64 = 64
	// versionOffset 版本号在文件中下标
	versionOffset = 0
	// blockLenConfigOffset  数据库在文件中下标
	blockLenConfigOffset = 8
	// offsetConfigOffset 偏移量在文件中的下标
	offsetConfigOffset = 16

	blockValidOffset = 0
	// defaultValidLen
	blockValidLen            = 1
	blockDataLenConfigOffset = blockValidOffset + blockValidLen
	// 数据长度字节长度
	blockDataLenConfigOffLen    = 8
	blockDataLenConfigOffsetEnd = blockDataLenConfigOffset + blockDataLenConfigOffLen
)

// 下面常量部分配置为默认值。并且有可能会写入文件中
const (
	libVersion = 1
	// 默认偏移量
	defaultOffset = 0
	// 默认数据块长度
	defaultBlockLen = 128
)

type QueueHeader struct {
	// 版本号
	version int64
	// 块长度
	blockLen int64
	// 偏移量，记录了下一个要出队数据的文件坐标
	offset int64
	// 数据最大长度 = 块长度 - 有效位长度 - 实际数据长度记录位长度
	dataMaxLen int64
}

type FileQueue struct {
	queueDir     string
	drLock       sync.Mutex
	queueHandle  *os.File
	header       *QueueHeader
	readInt64Buf [8]byte
}

// Vacuum 压缩文件，清理已经出队的数据
func (itself *FileQueue) Vacuum() error {
	itself.drLock.Lock()
	itself.drLock.Unlock()
	var err error
	tmpQueueHandle, err := OpenOrCreateFile(itself.getQueueTmpPath())
	if err != nil {
		return err
	}
	// 迁移头
	header := make([]byte, headLen)
	if _, err = itself.readAt(header, 0); err != nil {
		return err
	}
	if _, err = tmpQueueHandle.WriteAt(header, 0); err != nil {
		return err
	}
	mDataLen := 1024 * 1024
	blockData := make([]byte, mDataLen)
	var i int64
	// 迁移剩余队列
	for {
		lastN, _ := itself.readAt(blockData, itself.header.offset*itself.header.blockLen+headLen+i*int64(mDataLen))
		if lastN < mDataLen {
			// 如果获取的数据小于一个数据块儿，说明是最后一块。单独处理
			lastData := make([]byte, lastN)
			for di := 0; di < lastN; di++ {
				lastData[di] = blockData[di]
			}
			if _, err = tmpQueueHandle.WriteAt(lastData, headLen+i*int64(mDataLen)); err != nil {
				return err
			}
			break
		} else {
			if _, err = tmpQueueHandle.WriteAt(blockData, headLen+i*int64(mDataLen)); err != nil {
				return err
			}
		}

		i += 1
	}
	// 新队列重制偏移量
	itself.header.offset = defaultOffset
	_, err = tmpQueueHandle.WriteAt(Int64ToBytes(itself.header.offset), offsetConfigOffset)
	if err != nil {
		return err
	}
	_ = itself.queueHandle.Close()
	if err = os.Remove(itself.getQueuePath()); err != nil {
		return err
	}
	_ = tmpQueueHandle.Close()
	if err = os.Rename(itself.getQueueTmpPath(), itself.getQueuePath()); err != nil {
		return err
	}
	itself.queueHandle, err = os.OpenFile(itself.getQueuePath(), os.O_RDWR, 0666)

	if err != nil {
		return err
	}
	return nil
}

// getQueuePath 队列文件，当前只有一个文件，内容为 header + queueBlockList
func (itself *FileQueue) getQueuePath() string {
	return itself.queueDir + "/q.dat"
}

// getQueueTmpPath 获取队列临时目录，这个是用来清理消费时的新文件
func (itself *FileQueue) getQueueTmpPath() string {
	return itself.getQueuePath() + ".tmp"
}

// init 文件队列初始化函数
// 会检测是否存在队列仓库目录，没有的话进行创建，同时初始化队列文件的header
// 如果存在则读取上次的header ,header 中存在version ，当前队列下标信息
func (itself *FileQueue) init() error {
	var err error
	itself.queueHandle, err = OpenOrCreateFile(itself.getQueuePath())
	if err != nil {
		return err
	}
	headerData := make([]byte, headLen)
	n, err := itself.readAt(headerData, blockLenConfigOffset)
	if n == 0 {
		err = itself.writeHeader()
	} else {
		err = itself.readHeader()
		if err == nil {
			slog.Info("读取header成功")
		}
	}
	return err
}

// Len 队列有效长度，暂未实现
func (itself *FileQueue) Len() int64 {
	itself.drLock.Lock()
	defer itself.drLock.Unlock()
	n, _ := itself.queueHandle.Seek(0, io.SeekEnd)
	return (n - headLen) / itself.header.blockLen

}

// Push 入队
// 传入数据 ，拼接数据块 有效位 长度 真实数据位
// 追加至文件尾
func (itself *FileQueue) Push(data string) error {
	itself.drLock.Lock()
	defer itself.drLock.Unlock()
	// 有效表示位
	dataByte := []byte(data)
	if len(dataByte) > int(itself.header.dataMaxLen) {
		return errors.New("当前数据长度超过最大长度")
	}
	unitData := make([]byte, itself.header.blockLen)
	unitData[0] = 1
	//copy(unitData[1:], Int64ToBytes(int64(len(dataByte))))
	ReplaceData(unitData, Int64ToBytes(int64(len(dataByte))), 1)
	ReplaceData(unitData, dataByte, 9)
	n, _ := itself.queueHandle.Seek(0, io.SeekEnd)
	_, err := itself.writeAt(unitData, n)
	return err
}

// Pop 出队
// 计算偏移量 ，读取 数据块 ，读取长度位 ，读取对应长度数据
// 读取成功 设置最新的下标并写入文件
func (itself *FileQueue) Pop() (string, error) {
	itself.drLock.Lock()
	defer itself.drLock.Unlock()
	// 数据块起始位置 head + block * n
	blockOffset := itself.header.offset*itself.header.blockLen + headLen
	oData := make([]byte, itself.header.blockLen)
	if _, err := itself.readAt(oData, blockOffset); err != nil {
		return "", err
	}
	dataLen := BytesToInt64(oData[blockDataLenConfigOffset:blockDataLenConfigOffsetEnd])
	data := oData[blockDataLenConfigOffsetEnd : blockDataLenConfigOffsetEnd+dataLen]
	if err := itself.updateOffset(); err != nil {
		return "", err
	}
	return string(data), nil
}

// 更新偏移
func (itself *FileQueue) updateOffset() error {
	itself.header.offset += 1
	_, err := itself.writeInt64At(itself.header.offset, offsetConfigOffset)
	if err != nil {
		return err
	}
	return nil
}

// writeHeader 写入头信息
func (itself *FileQueue) writeHeader() error {
	data := make([]byte, 64)
	binary.LittleEndian.PutUint64(data[versionOffset:versionOffset+8], uint64(itself.header.version))
	binary.LittleEndian.PutUint64(data[blockLenConfigOffset:blockLenConfigOffset+8], uint64(itself.header.blockLen))
	binary.LittleEndian.PutUint64(data[offsetConfigOffset:offsetConfigOffset+8], uint64(itself.header.offset))
	binary.LittleEndian.PutUint64(data[24:64], 0)
	if _, err := itself.queueHandle.WriteAt(data, 0); err != nil {
		return err
	}
	return nil
}

func (itself *FileQueue) readHeader() error {
	data := make([]byte, 64)
	if _, err := itself.queueHandle.ReadAt(data, 0); err != nil {
		return err
	}
	version := binary.LittleEndian.Uint64(data[:8])
	blockLen := binary.LittleEndian.Uint64(data[8:16])
	offset := binary.LittleEndian.Uint64(data[16:24])
	itself.header.version = int64(version)
	itself.header.blockLen = int64(blockLen)
	itself.header.offset = int64(offset)
	itself.header.dataMaxLen = itself.header.blockLen - blockValidLen - blockDataLenConfigOffLen
	return nil
}

// write 在队列文件写入数据
func (itself *FileQueue) write(data []byte) (int, error) {
	return itself.queueHandle.Write(data)
}

// writeAt 在队列文件指定位置写入数据
func (itself *FileQueue) writeAt(data []byte, off int64) (int, error) {
	return itself.queueHandle.WriteAt(data, off)
}

// writeInt64At 在队列文件指定位置写入一个int64的数据
func (itself *FileQueue) writeInt64At(data int64, off int64) (int, error) {
	return itself.queueHandle.WriteAt(Int64ToBytes(data), off)
}

// readAt 在队列文件指定位置读取数据
func (itself *FileQueue) readAt(b []byte, off int64) (n int, err error) {
	return itself.queueHandle.ReadAt(b, off)
}

// readInt64At 在队列文件指定位置读取一个int64的数据
func (itself *FileQueue) readInt64At(off int64) (data int64, err error) {
	_, err = itself.readAt(itself.readInt64Buf[:], off)
	if err != nil {
		return
	}
	data = int64(binary.LittleEndian.Uint64(itself.readInt64Buf[:]))
	return
}

// util

func Int64ToBytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	if len(buf) < 8 {
		buf = append(make([]byte, 8-len(buf)), buf...)
	}
	return int64(binary.LittleEndian.Uint64(buf))
}

// ReplaceData 替换指定位置之后的数据
func ReplaceData(o []byte, d []byte, i int) {
	copy(o[i:], d)
}

func OpenOrCreateFile(path string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
}
