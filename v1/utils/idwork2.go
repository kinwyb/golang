package utils

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

//SnowFlakeID 唯一序号生成对象
//按毫秒生成，16位唯一编码
type SnowFlakeID struct {
	workerID           int         //机器ID
	twepoch            int64       //唯一时间，这是一个避免重复的随机量，自行设定不要大于当前时间戳
	sequence           uint32      //当前时间戳计数器
	workerIDBits       uint        //机器码字节数,默认4字节保存机器ID
	maxWorkerID        int64       //最大机器ID
	sequenceBits       uint        //计数器字节数,默认10个字节保存计数器
	workerIDShift      uint        //机器码数据左移位数
	timestampLeftShift uint        //时间戳左移位数
	sequenceMask       int         //一微秒内可以产生的计数器值,达到该值后要等到下一微妙再生成
	lastTimestamp      int64       //上次生成序号的时间戳
	lock               *sync.Mutex //同步锁
}

//NewSnowFlakeID 生成一个SnowFlakeID对象
func NewSnowFlakeID(workerID int, params ...int64) *SnowFlakeID {
	var sequenceBits uint
	var twepoch int64
	if len(params) > 1 {
		sequenceBits = uint(params[0])
		twepoch = params[1]
	} else {
		if len(params) > 0 {
			sequenceBits = uint(params[0])
			twepoch = 1497039792410
		} else {
			sequenceBits = 12
			twepoch = 1497039792410
		}
	}
	idw := &SnowFlakeID{
		workerID:     workerID,
		twepoch:      twepoch,
		sequenceBits: sequenceBits,
		sequence:     0,
		workerIDBits: 10,
		lock:         &sync.Mutex{},
	}
	idw.maxWorkerID = -1 ^ -1<<idw.workerIDBits
	idw.sequenceMask = -1 ^ -1<<idw.sequenceBits
	idw.workerIDShift = idw.sequenceBits
	idw.timestampLeftShift = idw.workerIDBits + idw.workerIDShift
	idw.lastTimestamp = -1
	return idw
}

//Next 生成一个唯一ID
func (d *SnowFlakeID) Next() (int64, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	timestamp := d.timeGen()
	if timestamp < d.lastTimestamp {
		return 0, fmt.Errorf("Clock moved backwards. Refusing to generate id for %d milliseconds", d.lastTimestamp)
	}
	if d.lastTimestamp == timestamp {
		atomic.AddUint32(&d.sequence, 1)
		d.sequence = uint32(int(d.sequence) & d.sequenceMask)
		if d.sequence == 0 {
			timestamp = d.tilNextMillis()
			d.lastTimestamp = timestamp
		}
	} else {
		d.sequence = 0
		d.lastTimestamp = timestamp
	}
	i := ((timestamp - d.twepoch) << d.timestampLeftShift) | int64((d.workerID << d.workerIDShift)) | int64(d.sequence)
	return i, nil
}

//NextString 下一个唯一字符串
func (d *SnowFlakeID) NextString() (string, error) {
	id, err := d.Next()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

//获取下一微秒时间戳
func (d *SnowFlakeID) tilNextMillis() int64 {
	timestamp := timeGen()
	for {
		if timestamp > d.lastTimestamp {
			break
		}
		timestamp = timeGen()
	}
	return timestamp
}

//当前时间戳
func (d *SnowFlakeID) timeGen() int64 {
	return time.Now().UnixNano() / 1e6
}
