package canstop

import (
	"container/ring"
	"math"
	"time"
)

type ErrorCounter struct {
	*ring.Ring
}

func NewErrorCounter(size int) *ErrorCounter {
	return &ErrorCounter{ring.New(size)}
}
func (self *ErrorCounter) AddTimestamp(t time.Time) {
	self.Ring.Value = &t
	self.Ring = self.Ring.Next()
}

func (self *ErrorCounter) CalculateDelay() (delay time.Duration) {
	rate := self.calculateRate(time.Now())
	power := math.Min(math.Ceil(rate)-1.0, 3)
	delay = time.Duration(math.Pow(2.0, power)) * time.Second
	return
}

func (self *ErrorCounter) calculateRate(upTo time.Time) (rate float64) {
	var (
		start *time.Time
		count int
	)
	self.Ring.Do(func(x interface{}) {
		if x != nil {
			if start == nil {
				start = x.(*time.Time)
			}
			count++
		}
	})

	if count == 0 {
		return
	}

	totalDuration := upTo.Sub(*start)
	// deal with the case of a single recent datapoint where totalDuration could be 0 otherwise
	if totalDuration < time.Second && count == 1 {
		totalDuration = time.Second
	}
	secDuration := float64(totalDuration) / float64(time.Second)

	rate = float64(count) / secDuration

	return
}
