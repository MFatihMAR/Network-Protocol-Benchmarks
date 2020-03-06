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
}

func NewCondition(cfg *Config) (*Condition, error) {
	// todo
	return nil, errors.New("not implemented")
}

func (c *Condition) Close() {
	c.closeOnce.Do(func() {
		c.closed = true

		// todo
	})
}
