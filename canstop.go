package canstop

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type PanicError struct {
	p     interface{}
	stack []byte
}

func (self PanicError) Error() string {
	return fmt.Sprintf("panic: '%s' at:\n%s", self.p, string(self.stack))
}

type Manageable func(t *Lifecycle) error

type Lifecycle struct {
	wg                  *sync.WaitGroup
	once                *sync.Once
	services            map[chan string]string
	interrupt           chan bool
	serviceRegistration chan chan string
}

func NewLifecycle() *Lifecycle {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	reg := make(chan chan string)
	services := make(map[chan string]string)
	// listen for services registering
	go func() {
		for {
			ch := <-reg
			services[ch] = <-ch
		}
	}()
	return &Lifecycle{wg, &sync.Once{}, services, make(chan bool), reg}
}

// Session allows a single session to run with the ability to check for
// cancellation.  Sessions are expected to be plentiful and to come and go many
// times during the course of a program's life time. They are not accounted for
// the same way that services are, but are nonetheless given a chance to clean
// up at shutdown time. A good example of a session is a single TCP connection.
func (self *Lifecycle) Session(f Manageable) {
	self.wg.Add(1)
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = PanicError{r, debug.Stack()}
		}
		if err != nil {
			log.Printf("Session ended in error: %s\n", err)
		}
		self.wg.Done()
	}()
	err = f(self)
}

// Service allows a background process that is expected to run for the duration
// of a program to run with the ability to check for cancellation. A good
// example of a service is the "accept loop" of a network service, which
// perpetually accepts incoming connections.
func (self *Lifecycle) Service(f Manageable, name string) {
	// we pass a channel back to the Lifecycle registration goroutine
	// over which we communicate the name of the registering service.
	// closure of this channel is used to indicate clean termination
	imFinished := make(chan string)
	self.serviceRegistration <- imFinished
	imFinished <- name
	self.wg.Add(1)
	defer self.wg.Done()
	defer close(imFinished)
	// services should be restarted if they stop running for any reason
	// hopefully f itself is a loop that is also reading from the interrupt
	// channel; we should only hit the top of this loop on errors/panics
	ec := NewErrorCounter(5)
	for !self.IsInterrupted() {
		time.Sleep(ec.CalculateDelay())
		err := loopCalmly(self, f, name)
		if err != nil {
			log.Printf("error in service %s: %s", name, err)
			ec.AddTimestamp(time.Now())
		}
	}
}

// Interrupt returns the channel which is closed to signal cancellation.
func (self *Lifecycle) Interrupt() <-chan bool {
	return self.interrupt
}

func (self *Lifecycle) StopAndWait() {
	self.Stop(math.MaxInt16 * time.Hour)
}

func waitOnWaitGroup(wg *sync.WaitGroup) (ch chan bool) {
	ch = make(chan bool)
	waiter := func(ch chan bool, wg *sync.WaitGroup) {
		wg.Wait()
		ch <- true
	}
	go waiter(ch, wg)
	return ch
}

func (self *Lifecycle) Stop(maxWait time.Duration) {
	log.Printf("Orderly shutdown commenced")
	self.once.Do(func() {
		self.stopBody(maxWait)
	})
}

// convenience method for checking for interrupt for non-select{} usecases
func (l *Lifecycle) IsInterrupted() bool {
	select {
	case <-l.Interrupt():
		{
			return true
		}
	default:
		{
			return false
		}
	}
}

func (self *Lifecycle) stopBody(maxWait time.Duration) {
	close(self.interrupt)
	self.wg.Done()
	waiter := waitOnWaitGroup(self.wg)
	select {
	case <-waiter:
		{
			return
		}
	case <-time.After(maxWait):
		{
			laggards := make([]string, 0)
			for finishedChan, service := range self.services {
				select {
				case _, closed := <-finishedChan:
					{
						if closed {
							continue
						}
					}
				default:
					{
						laggards = append(laggards, service)
					}
				}
			}
			if len(laggards) > 0 {
				log.Printf("The following services did not terminate in a timely fashion: %s\n", laggards)
			}
		}
	}
}

// loopCalmly is the body of a service loop with panic protection and error
// reporting
func loopCalmly(l *Lifecycle, f Manageable, name string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = PanicError{r, debug.Stack()}
		}
		runtime.Gosched() // break up panic hotloops
	}()
	err = f(l)
	return
}
