package utils

import "time"

// Clock - інтерфейс для роботи з часом, що дозволяє легко мокати його в тестах
type Clock interface {
	Now() time.Time
	NowUnixMilli() int64
	NowUnixNano() int64
}

type realClock struct{}

func NewRealClock() Clock {
	return &realClock{}
}

func (c *realClock) Now() time.Time {
	return time.Now()
}

func (c *realClock) NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

func (c *realClock) NowUnixNano() int64 {
	return time.Now().UnixNano()
}
