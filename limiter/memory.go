package limiter

import (
	"context"
	"sync"
	"time"
)

type counter struct {
	buckets     map[uint64]uint64
	duration    time.Duration
	granularity time.Duration
	locker      sync.Mutex
}

type SimpleCounterLimit struct {
	counter     map[string]*counter
	duration    time.Duration
	granularity time.Duration
	limit       uint64
}

func NewSimpleCounterLimit(limit uint64, duration, granularity time.Duration) *SimpleCounterLimit {
	return &SimpleCounterLimit{
		counter:     make(map[string]*counter),
		duration:    duration,
		limit:       limit,
		granularity: granularity,
	}
}

func (counter *counter) verify(limit uint64) (after time.Time, allowed bool) {
	counter.locker.Lock()
	defer counter.locker.Unlock()

	now := time.Now()

	window := uint64(now.Add(counter.duration*-1).UnixNano() / int64(counter.granularity))

	requests := uint64(0)
	for timestamp, count := range counter.buckets {
		if timestamp >= window {
			requests += count
		} else {
			// Remove old and unused buckets
			delete(counter.buckets, timestamp)
		}
	}

	if requests >= limit {
		return now.Add(counter.granularity), false
	}

	if _, found := counter.buckets[window]; !found {
		counter.buckets[window] = 0
	}

	counter.buckets[window] += 1

	return time.Time{}, true
}

func (s *SimpleCounterLimit) IsAllowed(ctx context.Context, token string) (time.Time, bool) {
	if _, found := s.counter[token]; !found {
		s.counter[token] = &counter{
			buckets:     make(map[uint64]uint64),
			duration:    s.duration,
			granularity: s.granularity,
		}
	}

	return s.counter[token].verify(s.limit)
}
