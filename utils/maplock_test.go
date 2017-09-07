package utils

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func Test_Lock(t *testing.T) {
	mlock := NewMapLock()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	a := 0
	b := 0
	c := 0
	g := &sync.WaitGroup{}
	for i := 1; i <= 100; i++ {
		g.Add(1)
		go func(i int) {
			mlock.Lock("a")
			time.Sleep(time.Duration(r.Intn(1000)) * time.Microsecond)
			a = a + i
			mlock.Unlock("a")
			g.Done()
		}(i)
	}
	for i := 1; i < 1000; i++ {
		go func(i int) {
			g.Add(1)
			mlock.Lock("b")
			time.Sleep(time.Duration(r.Intn(1000)) * time.Microsecond)
			b = b + i
			mlock.Unlock("b")
			g.Done()
		}(i)
	}
	for i := 1; i < 1000; i++ {
		go func(i int) {
			g.Add(1)
			mlock.Lock("c")
			time.Sleep(time.Duration(r.Intn(1000)) * time.Microsecond)
			c = c + i
			mlock.Unlock("c")
			g.Done()
		}(i)
	}
	g.Wait()
	println(a)
	println(b)
	println(c)
}
