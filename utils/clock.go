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

// MockClock - реалізація для тестів
type MockClock struct {
	FixedTime time.Time
}

func NewMockClock(t time.Time) *MockClock {
	return &MockClock{FixedTime: t}
}

func (c *MockClock) Now() time.Time {
	return c.FixedTime
}

func (c *MockClock) NowUnixMilli() int64 {
	return c.FixedTime.UnixMilli()
}

func (c *MockClock) NowUnixNano() int64 {
	return c.FixedTime.UnixNano()
}

func (c *MockClock) Add(d time.Duration) {
	c.FixedTime = c.FixedTime.Add(d)
}
