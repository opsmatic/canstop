package canstop

import (
	"launchpad.net/tomb"
	"log"
	"sync"
	"time"
)

type Graceful interface {
	Run(t *tomb.Tomb)
}

type Runner interface {
	RunMe(g Graceful)
	Stop()
	Wait()
}

func NewRunner(maxWait time.Duration) Runner {
	r := &runner{
		&sync.WaitGroup{}, make([]*tomb.Tomb, 0), maxWait,
	}
	// avoid race: Wait() being called before jobs have had a chance to Add()
	r.Add(1)
	return r
}

type runner struct {
	*sync.WaitGroup
	jobs    []*tomb.Tomb
	maxWait time.Duration
}

// This could just take func(*Tomb) but an interface feels cleaner
func (self *runner) RunMe(g Graceful) {
	t := &tomb.Tomb{}
	self.jobs = append(self.jobs, t)
	go func() {
		defer markDone(t)
		g.Run(t)
	}()
}

func markDone(t *tomb.Tomb) {
	select {
	case _ = <-t.Dead():
		{
			// do nothing because that means the Graceful took care to call .Done()
		}
	default:
		{
			t.Done()
		}
	}
}

func stopJob(t *tomb.Tomb, timeout time.Duration, group *sync.WaitGroup) {
	t.Kill(nil)
	timeoutChan := time.After(timeout)
	select {
	case _ = <-t.Dead():
		{
			markDone(t)
		}
	case _ = <-timeoutChan:
		{
			t.Killf("Job took too long to terminate, forcing termination after %d\n", timeout)
			t.Done()
		}
	}
	group.Done()
}

func (self *runner) Stop() {
	self.Done()
	for _, t := range self.jobs {
		self.Add(1)
		go stopJob(t, self.maxWait, self.WaitGroup)
	}
	for _, t := range self.jobs {
		<-t.Dead()
		if err := t.Err(); err != nil {
			log.Printf("Ungraceful stop: %s\n", err)
		}
	}
}
