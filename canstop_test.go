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
	defer t.Done()
	for {
		select {
		case _ = <-t.Dying():
			{
				return
			}
		default:
		}
		self.counter++
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *MySuite) TestRunner(c *C) {
	// really rough timing-based test, but deal with it
	r := NewRunner()
	g := &testGraceful{0}
	r.RunMe(g)
	time.Sleep(500 * time.Millisecond)
	r.Stop()
	c.Check(g.counter > 0, Equals, true)
	c.Check(g.counter < 10, Equals, true)
}
