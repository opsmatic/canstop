package canstop

import (
	"log"
	"math"
	"sync"
	"time"
)

type Manageable func(t Lifecycle) error

type Lifecycle interface {
	ManageService(Manageable, string)
	ManageSession(Manageable)
	Interrupt() chan bool
	StopAndWait()
	Stop(time.Duration)
}

/**
 * convenience method for checking for interrupt for non-select{} usecases
 */
func IsInterrupted(l Lifecycle) bool {
	select {
	case <-l.Interrupt():
		return true
	default:
		return false
	}
}

func NewLifecycle() Lifecycle {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return &lifecycle{wg, &sync.Once{}, make(map[chan bool]string), make(chan bool)}
}

type lifecycle struct {
	*sync.WaitGroup
	*sync.Once
	services  map[chan bool]string
	interrupt chan bool
}

func (self *lifecycle) ManageSession(f Manageable) {
	self.Add(1)
	defer self.Done()
	err := f(self)
	if err != nil {
		log.Printf("Session ended uncleanly: %s\n")
	}
}

func (self *lifecycle) ManageService(f Manageable, name string) {
	imFinished := make(chan bool)
	self.services[imFinished] = name
	self.Add(1)
	defer self.Done()
	var err error
	// services should be restarted if they stop running for any reason
	// hopefully f itself is a loop that is also reading from the interrupt
	// channel; we should only hit the top of this loop on errors/panics
	for !IsInterrupted(self) {
		err = f(self)
		if err != nil {
			log.Printf("Service %s crashed with error: %s", name, err)
		}
	}
	if err != nil {
		log.Printf("Service %s exited with an error: %s", name, err)
	}
	close(imFinished)
}

func (self *lifecycle) Interrupt() chan bool {
	return self.interrupt
}

func (self *lifecycle) StopAndWait() {
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

func (self *lifecycle) Stop(maxWait time.Duration) {
	self.Once.Do(func() {
		self.stopBody(maxWait)
	})
}

func (self *lifecycle) stopBody(maxWait time.Duration) {
	close(self.interrupt)
	self.WaitGroup.Done()
	waiter := waitOnWaitGroup(self.WaitGroup)
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
				case <-finishedChan:
					continue
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
