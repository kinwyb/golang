package utils

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func Test_SnowFlakeID(t *testing.T) {
	idWorker := NewSnowFlakeID(10, time.Nanosecond)
	lock := &sync.RWMutex{}
	m := map[string]string{}
	count := int64(0)
	w := &sync.WaitGroup{}
	convey.Convey("ID生成器", t, func() {
		for a := 0; a <= 50; a++ {
			w.Add(1)
			go func() {
				var id string
				var err error
				for i := 0; i < 100000; i++ {
					id, err = idWorker.NextString()
					if err != nil {
						convey.Printf("错误:%s\n", err.Error())
						w.Done()
						convey.So(false, convey.ShouldBeTrue)
						return
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
		convey.So(count, convey.ShouldEqual, 0)
	})
}
