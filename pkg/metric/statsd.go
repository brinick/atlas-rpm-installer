package metric

import (
	"fmt"
)

// API provides the set of methods for gathering metrics
type API interface {
	Count(string, int) string
	Gauge() string
	Timer() string
}

// NewQueue returns a new metric queue
func NewQueue(itemFormatter string, cap int) *Queue {
	if cap <= 0 {
		cap = 1
	} else if cap >= 100 {
		cap = 100
	}

	switch itemFormatter {
	case "statsd":
		return &Queue{
			formatter: &StatsD{},
			queue:     make(chan string, cap),
		}
	default:
		panic(fmt.Sprintf("unknown metric formatter %s", itemFormatter))
	}
}

// Queue is a channel of metric strings, encoded by the given formatter
type Queue struct {
	formatter API
	queue     chan string
}

// Items returns the channel of metric items
func (mq *Queue) Items() <-chan string {
	return mq.queue
}

// Count is
func (mq *Queue) Count(name string, value int) {
	if mq != nil {
		mq.queue <- mq.formatter.Count(name, value)
	}
}

// Gauge is
func (mq *Queue) Gauge() {
	if mq != nil {
		mq.queue <- mq.formatter.Gauge()
	}
}

// Timer is
func (mq *Queue) Timer() {
	if mq != nil {
		mq.queue <- mq.formatter.Timer()
	}
}

// ------------------------------------------------------------------

// StatsD provides the format for a StatsD metrics gatherer
type StatsD struct {
}

// Count provides the format for StatsD Count
func (s *StatsD) Count(name string, value int) string {
	return fmt.Sprintf("")
}

// Gauge provides the format for StatsD Gauge
func (s *StatsD) Gauge() string {
	return fmt.Sprintf("")
}

// Timer provides the format for StatsD Timer
func (s *StatsD) Timer() string {
	return fmt.Sprintf("")
}
