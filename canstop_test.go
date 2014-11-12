package canstop

import (
	"sync/atomic"
	"testing"
	"time"

	. "launchpad.net/gocheck"
)

// boilerplate
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

type testGraceful struct {
	counter int
}

func (self *testGraceful) Run(l *Lifecycle) error {
	for !l.IsInterrupted() {
		self.counter++
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

type testUnGraceful struct{}

func (self *testUnGraceful) Run(l *Lifecycle) error {
	for {
		time.Sleep(5 * time.Hour)
	}
	return nil
}

func (s *MySuite) TestRunner(c *C) {
	// really rough timing-based test, but deal with it
	m := NewLifecycle()
	g := &testGraceful{0}
	go m.Service(g.Run, "graceful")
	time.Sleep(500 * time.Millisecond)
	m.Stop(100 * time.Millisecond)
	c.Check(g.counter > 3, Equals, true)
	c.Check(g.counter < 6, Equals, true)
}

func (s *MySuite) TestTimeout(c *C) {
	m := NewLifecycle()
	g := &testUnGraceful{}
	go m.Service(g.Run, "ungraceful")
	done := make(chan bool)
	go func() {
		m.Stop(100 * time.Millisecond)
		done <- true
	}()
	timeout := time.After(150 * time.Millisecond)
	select {
	case _ = <-done:
		{
		}
	case _ = <-timeout:
		{
			c.Fail()
		}
	}
}
func (s *MySuite) TestDoubleStop(c *C) {
	l := NewLifecycle()

	// the second call will panic if the protection isn't set correctly
	l.StopAndWait()
	l.StopAndWait()
}

func (s *MySuite) TestPanicSession(c *C) {
	l := NewLifecycle()

	var counter int32 = 0

	go l.Session(func(l *Lifecycle) (e error) {
		for i := 0; i < 10; i++ {
			atomic.AddInt32(&counter, 1)
			time.Sleep(10 * time.Millisecond)
		}
		return
	})
	go l.Session(func(l *Lifecycle) (e error) {
		panic("Immediate panic, oh god!!")
		return
	})

	l.Stop(100 * time.Millisecond)
	c.Check(atomic.LoadInt32(&counter) > 5, Equals, true)
}

func (s *MySuite) TestPanicService(c *C) {
	l := NewLifecycle()

	var counter int32 = 0
	go l.Service(func(l *Lifecycle) (e error) {
		for i := 0; !l.IsInterrupted() && i < 10; i++ {
			shouldPanic := (atomic.LoadInt32(&counter) == 5)
			atomic.AddInt32(&counter, 1)
			if shouldPanic {
				panic("Let's interrupt this story")
			}
		}
		return
	}, "panicker")
	time.Sleep(100 * time.Millisecond)
	l.StopAndWait()
	c.Check(atomic.LoadInt32(&counter) > 6, Equals, true)
}
