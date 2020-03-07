package condition

import (
	"errors"
	"sync"
)

type Condition struct {
	Err error

	closeOnce sync.Once
	closed    bool
}

type Config struct {
	LatencyMsMax  int   // todo: explain
	LatencyMsMin  int   // todo: explain
	LossPer1kMax  int   // todo: explain
	LossPer1kMin  int   // todo: explain
	BandwidthMax  int   // todo: explain
	BandwidthMin  int   // todo: explain
	MTUs          []int // todo: explain
	UpdateRateSec int   // todo: explain
}

func NewCondition(cfg *Config) (*Condition, error) {
	// todo
	return nil, errors.New("not implemented")
}

func (c *Condition) UseBandwidth(size int) bool {
	// todo
	return false
}

func (c *Condition) CheckLoss() bool {
	// todo
	return false
}

func (c *Condition) AddLatency() int {
	// todo
	return 0
}

func (c *Condition) Close() {
	c.closeOnce.Do(func() {
		c.closed = true

		// todo
	})
}
