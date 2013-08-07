package canstop

import (
	. "launchpad.net/gocheck"
	"launchpad.net/tomb"
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

func (self *testGraceful) Run(t *tomb.Tomb) {
	for {
		select {
		case _ = <-t.Dying():
			{
				t.Done()
				return
			}
		default:
		}
		self.counter++
		time.Sleep(100 * time.Millisecond)
	}
}

type testUnGraceful struct{}

func (self *testUnGraceful) Run(t *tomb.Tomb) {
	for {
		time.Sleep(5 * time.Hour)
	}
}

func (s *MySuite) TestRunner(c *C) {
	// really rough timing-based test, but deal with it
	r := NewRunner(5 * time.Second)
	g := &testGraceful{0}
	r.RunMe(g)
	time.Sleep(500 * time.Millisecond)
	r.Stop()
	c.Check(g.counter > 3, Equals, true)
	c.Check(g.counter < 6, Equals, true)
}

func (s *MySuite) TestTimeout(c *C) {
	r := NewRunner(100 * time.Millisecond)
	g := &testUnGraceful{}
	r.RunMe(g)
	r.Stop()
	done := make(chan bool)
	go func() {
		r.Wait()
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
