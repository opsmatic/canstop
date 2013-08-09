package canstop

import (
	"log"
	"math"
	"runtime"
	"sync"
	"time"
)

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

func (self *Lifecycle) Session(f Manageable) {
	self.wg.Add(1)
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Session ended in panic: %s\n", r)
		}
		if err != nil {
			log.Printf("Session ended in error: %s\n", err)
		}
		self.wg.Done()
	}()
	err = f(self)
}

func loopCalmly(l *Lifecycle, f Manageable, name string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Service %s panicked: %s\n", name, r)
		}
		runtime.Gosched() // break up panic hotloops
	}()
	err := f(l)
	if err != nil {
		log.Printf("Service %s errored: %s\n", name, err)
	}
}

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
	for !self.IsInterrupted() {
		loopCalmly(self, f, name)
	}
}

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
					if closed {
						continue
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

/**
 * convenience method for checking for interrupt for non-select{} usecases
 */
func (l *Lifecycle) IsInterrupted() bool {
	select {
	case <-l.Interrupt():
		return true
	default:
		return false
	}
}
