package canstop

import (
	"time"

	. "launchpad.net/gocheck"
)

type ErrorCounterSuite struct{}

var _ = Suite(&ErrorCounterSuite{})

func (self *ErrorCounterSuite) TestCalculateRateEmpty(c *C) {
	ec := NewErrorCounter(5)
	rateNone := ec.calculateRate(time.Now())
	c.Assert(rateNone, Equals, 0.0)
}

func (self *ErrorCounterSuite) TestCalculateRateSimple(c *C) {
	ec := NewErrorCounter(5)
	start := time.Now()
	ec.AddTimestamp(start)
	ec.AddTimestamp(start.Add(1 * time.Second))
	rate := ec.calculateRate(start.Add(1 * time.Second))
	c.Assert(rate, Equals, 2.0)
}
func (self *ErrorCounterSuite) TestCalculateRateMoreThanSizeValues(c *C) {
	ec := NewErrorCounter(5)
	start := time.Now()
	for i := 0; i < 10; i++ {
		ec.AddTimestamp(start.Add(time.Duration(i) * time.Second))
	}
	rate := ec.calculateRate(start.Add(10 * time.Second))
	c.Assert(rate, Equals, 1.0)
}

func (self *ErrorCounterSuite) TestCalculateRateAfterLull(c *C) {
	ec := NewErrorCounter(5)
	start := time.Now()
	for i := 0; i < 10; i++ {
		ec.AddTimestamp(start.Add(time.Duration(i) * time.Second))
	}
	rate := ec.calculateRate(start.Add(10 * time.Hour))
	c.Assert(rate < 0.01, Equals, true)
}

func (self *ErrorCounterSuite) TestCalculateRateStampede(c *C) {
	ec := NewErrorCounter(5)
	start := time.Now()
	for i := 0; i < 10; i++ {
		ec.AddTimestamp(start.Add(time.Millisecond))
	}
	rate := ec.calculateRate(start.Add(20 * time.Millisecond))
	c.Assert(rate > 2, Equals, true)
}

func (self *ErrorCounterSuite) TestCalculateRateSingleError(c *C) {
	ec := NewErrorCounter(5)
	start := time.Now()
	ec.AddTimestamp(start)
	rate := ec.calculateRate(start)
	c.Assert(rate > 0, Equals, true)
}

func (self *ErrorCounterSuite) TestCalculateDelayNoErrors(c *C) {
	ec := NewErrorCounter(5)
	c.Assert(ec.CalculateDelay(), Equals, time.Duration(0))
}

func (self *ErrorCounterSuite) TestCalculateDelayOneError(c *C) {
	ec := NewErrorCounter(5)
	start := time.Now()
	ec.AddTimestamp(start)
	c.Assert(ec.CalculateDelay() > 0, Equals, true)
	c.Assert(ec.CalculateDelay() < 5*time.Second, Equals, true)
}
func (self *ErrorCounterSuite) TestCalculateDelayStampedeError(c *C) {
	ec := NewErrorCounter(5)
	for i := 0; i < 10; i++ {
		ec.AddTimestamp(time.Now())
		time.Sleep(5 * time.Millisecond)
	}

	c.Assert(ec.CalculateDelay(), Equals, 8*time.Second)
}
