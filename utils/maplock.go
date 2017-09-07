package utils

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego"
)

//CountMutexLock 带计数器的同步锁
type CountMutexLock struct {
	sync.Mutex
	count   int32
	useTime int64
}

//Lock 加锁
func (c *CountMutexLock) Lock() {
	atomic.AddInt32(&c.count, 1)
	c.useTime = time.Now().Unix()
	c.Mutex.Lock()
}

//Unlock 解锁
func (c *CountMutexLock) Unlock() {
	c.useTime = time.Now().Unix()
	atomic.AddInt32(&c.count, -1)
	c.Mutex.Unlock()
}

type MapLock struct {
	m    map[string]*CountMutexLock
	lock *sync.RWMutex
}

func NewMapLock() *MapLock {
	m := &MapLock{
		m:    make(map[string]*CountMutexLock),
		lock: &sync.RWMutex{},
	}
	go m.del()
	return m
}

func (m *MapLock) Lock(key string) {
	m.lock.Lock()
	if l, ok := m.m[key]; ok {
		m.lock.Unlock()
		l.Lock()
		return
	}
	ret := &CountMutexLock{}
	m.m[key] = ret
	m.lock.Unlock()
	ret.Lock()
}

func (m *MapLock) Unlock(key string) {
	m.lock.RLock()
	if l, ok := m.m[key]; ok {
		m.lock.RUnlock()
		l.Unlock()
		return
	}
	m.lock.RUnlock()
}

func (m *MapLock) del() {
	defer func() {
		if err := recover(); err != nil {
			beego.BeeLogger.Error("自动清锁线程异常:%v", err)
			go m.del() //重启线程
		}
	}()
	for {
		t := time.Now().Unix()
		for k, v := range m.m {
			if t-v.useTime > 3600 && v.count < 1 {
				m.lock.Lock()
				delete(m.m, k)
				m.lock.Unlock()
			}
		}
		<-time.After(10 * time.Minute)
	}
}
