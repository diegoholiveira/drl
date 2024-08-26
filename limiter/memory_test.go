package limiter_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/diegoholiveira/drl/limiter"
)

func TestLimiterIsBlocking(t *testing.T) {
	limiter := limiter.NewSimpleCounterLimit(1, 30*time.Second, 500*time.Millisecond)

	_, allowed := limiter.IsAllowed("test")
	assert.True(t, allowed)

	_, allowed = limiter.IsAllowed("test")
	assert.False(t, allowed)

	time.Sleep(500 * time.Millisecond)

	_, allowed = limiter.IsAllowed("test")
	assert.True(t, allowed)
}

func TestRetryAfter(t *testing.T) {
	limiter := limiter.NewSimpleCounterLimit(1, 30*time.Second, 500*time.Millisecond)

	_, allowed := limiter.IsAllowed("test")
	assert.True(t, allowed)

	safeToRetry, allowed := limiter.IsAllowed("test")
	assert.False(t, allowed)

	for time.Now().Before(safeToRetry) {
	}

	_, allowed = limiter.IsAllowed("test")
	assert.True(t, allowed)
}
