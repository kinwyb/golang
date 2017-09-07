package utils

import (
	"sync"
	"sync/atomic"
	"testing"
)

func Test_NextID(t *testing.T) {
	idWorker := IDWorker(10)
	lock := &sync.RWMutex{}
	m := map[string]string{}
	count := int64(0)
	w := &sync.WaitGroup{}
	for a := 0; a <= 50; a++ {
		w.Add(1)
		go func() {
			var id string
			var err error
			for i := 0; i < 100000; i++ {
				id, err = idWorker.NextString()
				if err != nil {
					println("错误:" + err.Error())
					continue
				}
				lock.Lock()
				if _, ok := m[id]; ok {
					atomic.AddInt64(&count, 1)
				} else {
					m[id] = id
				}
				lock.Unlock()
			}
			w.Done()
		}()
	}
	w.Wait()
	if count > 0 {
		t.Errorf("重复次数:%d", count)
	}
}
