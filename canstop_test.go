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

func (self *testGraceful) Run(c *Control) {
	for !c.IsPoisoned() {
		self.counter++
		time.Sleep(100 * time.Millisecond)
	}
}

type testUnGraceful struct{}

func (self *testUnGraceful) Run(c *Control) {
	for {
		time.Sleep(5 * time.Hour)
	}
}

func (s *MySuite) TestRunner(c *C) {
	// really rough timing-based test, but deal with it
	r := NewManager(5 * time.Second)
	g := &testGraceful{0}
	r.Manage(g.Run, "graceful")
	time.Sleep(500 * time.Millisecond)
	r.Stop()
	c.Check(g.counter > 3, Equals, true)
	c.Check(g.counter < 6, Equals, true)
}

func (s *MySuite) TestTimeout(c *C) {
	r := NewManager(100 * time.Millisecond)
	g := &testUnGraceful{}
	r.Manage(g.Run, "ungraceful")
	done := make(chan bool)
	go func() {
		r.Stop()
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
