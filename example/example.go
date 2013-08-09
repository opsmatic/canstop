package main

import (
	"github.com/opsmatic/canstop"
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

func (self *Worker) Run(l *canstop.Lifecycle) error {
	for !l.IsInterrupted() {
		// do the important work
		work := <-self.Work
		if math.Mod(float64(work), float64(4)) == 0 {
			log.Printf("Found another multiple of 4! %d\n", work)
			self.found++
		}
		time.Sleep(1 * time.Second)
	}
	log.Printf("Orderly shutdown of Worker. Found %d matches\n", self.found)
	return nil
}

type Producer struct {
	Work chan int64
}

func (self *Producer) Run(l *canstop.Lifecycle) error {
	for {
		self.Work <- rand.Int63()
	}
	return nil
}

func main() {
	l := canstop.NewLifecycle()
	work := make(chan int64)
	w := NewWorker(work)
	p := &Producer{work}

	go l.Service(w.Run, "worker")
	go l.Service(p.Run, "producer")

	// run until we get a signal to stop
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	// send cancellation to all our jobs and wait on them to complete
	l.Stop(5 * time.Second)
}
