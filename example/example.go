package main

import (
	"github.com/opsmatic/canstop"
	"launchpad.net/tomb"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

type Worker struct {
	Work  chan int64
	found int64
}

func NewWorker(ch chan int64) *Worker {
	return &Worker{ch, 0}
}

func (self *Worker) Run(t *tomb.Tomb) {
	for {
		select {
		case _ = <-t.Dying():
			{
				log.Printf("Clean shut down of worker. Found %d matches\n", self.found)
				t.Done()
				return
			}
		default:
		}
		// do the important work
		work := <-self.Work
		if math.Mod(float64(work), float64(4)) == 0 {
			log.Printf("Found another multiple of 4! %d\n", work)
			self.found++
		}
		time.Sleep(1 * time.Second)
	}
}

type Producer struct {
	Work chan int64
}

func (self *Producer) Run(t *tomb.Tomb) {
	for {
		self.Work <- rand.Int63()
	}
}

func main() {
	r := canstop.NewRunner(5 * time.Second)
	work := make(chan int64)
	w := NewWorker(work)
	p := &Producer{work}

	r.RunMe(w)
	r.RunMe(p)

	// run until we get a signal to stop
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	// send cancellation to all our jobs and wait on them to complete
	r.Stop()
	r.Wait()
}
