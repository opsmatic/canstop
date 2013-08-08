package canstop

import (
	"launchpad.net/tomb"
	"log"
	"sync"
	"time"
)

// a service unit managed by Manager
type Managed func(t *Control)

/**
 * The singleton that manages the lifecycles of service units
 */
type Manager interface {
	Manage(f Managed, name string)
	Stop()
	Poison() chan bool
}

/**
 * Create a new instance of a Manager with the specified maximum time allowed
 * for service units to clean up after themselves at shutdown
 */
func NewManager(maxWait time.Duration) Manager {
	r := &manager{
		&sync.WaitGroup{}, make(map[*tomb.Tomb]string), maxWait, make(chan bool), &sync.Once{},
	}
	// avoid race: Wait() being called before jobs have had a chance to Add()
	r.Add(1)
	return r
}

type manager struct {
	*sync.WaitGroup
	jobs    map[*tomb.Tomb]string
	maxWait time.Duration
	poison  chan bool
	once    *sync.Once
}

// Manage a service unit; name is used for accounting
func (self *manager) Manage(f Managed, name string) {
	c := NewControl(self.WaitGroup, self.poison)
	self.jobs[c.Tomb] = name
	go func() {
		defer markDone(c.Tomb)
		f(c)
	}()
}

/**
 * a little convenience method to get around tomb.Done() panicking if called
 * twice
 */
func markDone(t *tomb.Tomb) {
	select {
	case _ = <-t.Dead():
		{
			// do nothing because that means the Managed took care to call .Done()
		}
	default:
		{
			t.Done()
		}
	}
}

/**
 * stop a job, giving it some time to clean up
 */
func stopJob(t *tomb.Tomb, timeout time.Duration, group *sync.WaitGroup) {
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

/**
 * Stop and wait
 */
func (self *manager) Stop() {
	self.once.Do(self.stopBody)
}

func (self *manager) stopBody() {
	close(self.poison)
	for t := range self.jobs {
		self.Add(1)
		go stopJob(t, self.maxWait, self.WaitGroup)
	}
	for t, name := range self.jobs {
		<-t.Dead()
		if err := t.Err(); err != nil {
			log.Printf("Ungraceful stop for service %s: %s\n", name, err)
		}
	}
	self.Done()
	self.Wait()
}

func (self *manager) Poison() chan bool {
	return self.poison
}
