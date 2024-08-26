package limiter

import (
	"context"
	"log"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client      *redis.Client
	duration    time.Duration
	granularity time.Duration
	limit       uint64
}

var script = redis.NewScript(`
local function sliding_window_ratelimit(KEYS, ARGV)
    local key = KEYS[1]
    local now = tonumber(ARGV[1])
    local window = tonumber(ARGV[2])
    local limit = tonumber(ARGV[3])
    local window_start = now - window

    redis.call('ZREMRANGEBYSCORE', key, 0, window_start)    -- O(log(N)+M)

    local count = redis.call('ZCARD', key)                  -- O(1)

    if count >= limit then
        return 0
	else
        redis.call('ZADD', key, now, now)                   -- O(log(N))
        redis.call('EXPIRE', key, math.ceil(window / 1000)) -- O(1)
        return 1
    end
end

return sliding_window_ratelimit(KEYS, ARGV)
`)

func NewRedisLimiter(client *redis.Client, limit uint64, duration, granularity time.Duration) *RedisLimiter {
	return &RedisLimiter{
		client:      client,
		duration:    duration,
		limit:       limit,
		granularity: granularity,
	}
}

func (s *RedisLimiter) IsAllowed(ctx context.Context, token string) (time.Time, bool) {
	now := time.Now()
	window := s.duration.Milliseconds()

	allowed, err := script.Run(ctx, s.client, []string{token}, now.UnixNano()/int64(time.Millisecond), window, s.limit).Bool()
	if err != nil {
		log.Printf("Error while scripting on Redis: %v", err)
		return time.Time{}, false
	}

	if allowed {
		return time.Time{}, true
	}

	return now.Add(s.granularity), false
}
