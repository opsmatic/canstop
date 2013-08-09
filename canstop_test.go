package canstop

import (
	. "launchpad.net/gocheck"
	"testing"
	"time"
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
