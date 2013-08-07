package canstop

import (
	"launchpad.net/tomb"
	"sync"
)

/**
 * The structure that gets passed to Managed instances so that the can know
 * when to shut down. It embeds a Tomb so that they can also record unfortunate
 * outcomes
 */
type Control struct {
	*tomb.Tomb
	wg *sync.WaitGroup

	//the poison pill channel that gets closed to signal shutdown
	Poison chan bool
}

func NewControl(wg *sync.WaitGroup, poison chan bool) *Control {
	return &Control{
		&tomb.Tomb{}, wg, poison,
	}
}

/**
 * run a simple unit of work, making sure that it has a chance to complete
 * after shut down is signaled. This should only be used for simple, short,
 * one-shot jobs. See Run() for longer-running goroutines, such as a
 * listen/accept loop or a queue consumer
 */
func (self *Control) RunSimple(f func()) {
	go func() {
		self.wg.Add(1)
		f()
		self.wg.Done()
	}()
}

/**
 * Run a function (presumably a loop), passing self as an argument, giving
 * access to the Poison channel. Thus a long running loop, such as a listen/
 * accept loop, can check the Poison channel between accepting connections.
 */
func (self *Control) Run(f func(c *Control)) {
	self.RunSimple(func() {
		f(self)
	})
}

/**
 * Used to check if shutdown has been requested; intended for loops that are
 * not selecting from multiple goroutines, such as listen/accept loops
 */
func (self *Control) IsPoisoned() bool {
	select {
	case <-self.Poison:
		return true
	default:
		return false
	}
}
