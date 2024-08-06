package g

import (
	"container/list"
	"runtime"
	"sync"

	"github.com/shanyux/leaf/conf"
	"github.com/shanyux/leaf/log"
)

// one FastGo per goroutine (goroutine not safe)
type FastGo struct {
	ChanCb    chan func()
	pendingGo int
}

type LinearGo struct {
	f  func()
	cb func()
}

type LinearContext struct {
	g              *FastGo
	linearGo       *list.List
	mutexLinearGo  sync.Mutex
	mutexExecution sync.Mutex
}

func New(l int) *FastGo {
	g := new(FastGo)
	g.ChanCb = make(chan func(), l)
	return g
}

func (g *FastGo) GoToExec(f func(), cb func()) {
	g.pendingGo++

	go func() {
		defer func() {
			g.ChanCb <- cb
			if r := recover(); r != nil {
				if conf.LenStackBuf > 0 {
					buf := make([]byte, conf.LenStackBuf)
					l := runtime.Stack(buf, false)
					log.Error("%v: %s", r, buf[:l])
				} else {
					log.Error("%v", r)
				}
			}
		}()

		f()
	}()
}

func (g *FastGo) Cb(cb func()) {
	defer func() {
		g.pendingGo--
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()

	if cb != nil {
		cb()
	}
}

func (g *FastGo) Close() {
	for g.pendingGo > 0 {
		g.Cb(<-g.ChanCb)
	}
}

func (g *FastGo) Idle() bool {
	return g.pendingGo == 0
}

func (g *FastGo) NewLinearContext() *LinearContext {
	c := new(LinearContext)
	c.g = g
	c.linearGo = list.New()
	return c
}

func (c *LinearContext) GoToExec(f func(), cb func()) {
	c.g.pendingGo++

	c.mutexLinearGo.Lock()
	c.linearGo.PushBack(&LinearGo{f: f, cb: cb})
	c.mutexLinearGo.Unlock()

	go func() {
		c.mutexExecution.Lock()
		defer c.mutexExecution.Unlock()

		c.mutexLinearGo.Lock()
		e := c.linearGo.Remove(c.linearGo.Front()).(*LinearGo)
		c.mutexLinearGo.Unlock()

		defer func() {
			c.g.ChanCb <- e.cb
			if r := recover(); r != nil {
				if conf.LenStackBuf > 0 {
					buf := make([]byte, conf.LenStackBuf)
					l := runtime.Stack(buf, false)
					log.Error("%v: %s", r, buf[:l])
				} else {
					log.Error("%v", r)
				}
			}
		}()

		e.f()
	}()
}
