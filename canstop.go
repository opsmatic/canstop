package canstop

import (
	"launchpad.net/tomb"
)

type Graceful interface {
	Run(t *tomb.Tomb)
}

type Runner interface {
	RunMe(g Graceful)
	Stop()
}

func NewRunner() Runner {
	return &runner{
		make([]*tomb.Tomb, 0),
	}
}

type runner struct {
	jobs []*tomb.Tomb
}

// This could just take func(*Tomb) but an interface feels cleaner
func (self *runner) RunMe(g Graceful) {
	t := &tomb.Tomb{}
	self.jobs = append(self.jobs, t)
	go func() {
		defer t.Done()
		g.Run(t)
	}()
}

func (self *runner) Stop() {
	for _, t := range self.jobs {
		t.Kill(nil)
		<-t.Dead()
	}
}
