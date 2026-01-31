package weir

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidBurst       = errors.New("config.Burst is below or equal 0")
	ErrInvalidCleanupRate = errors.New("config.CleanupRate should be a power of two")
	ErrInvalidKeyTTL      = errors.New("config.KeyTTL is below or equal 0")
	ErrInvalidRate        = errors.New("config.Rate is below or equal 0")
)

type RateLimiterOptions struct {
	Rate        time.Duration
	Burst       int
	KeyTTL      time.Duration
	CleanupRate time.Duration
	Shards      int

	shards     uint64
	burst      int64
	supplyRate int64
}

type Limiter struct {
	Config RateLimiterOptions
	shards []shard
}

func New(ctx context.Context, config RateLimiterOptions) (*Limiter, error) {
	if config.Burst <= 0 {
		return nil, ErrInvalidBurst
	}

	if config.CleanupRate <= 0 {
		return nil, ErrInvalidCleanupRate
	}

	if config.KeyTTL <= 0 {
		return nil, ErrInvalidKeyTTL
	}

	if config.Rate <= 0 {
		return nil, ErrInvalidRate
	}

	if !isPowerOfTwo(config.Shards) {
		config.Shards = toNearestPowerOfTwo(float64(config.Shards))
	}

	lim := &Limiter{
		Config: config,
		shards: make([]shard, config.Shards),
	}

	for i := range config.Shards {
		lim.shards[i] = shard{
			buckets: make(map[uint64]*bucket),
		}
	}

	lim.Config.shards = uint64(config.Shards)
	lim.Config.burst = int64(config.Burst)
	lim.Config.supplyRate = config.Rate.Nanoseconds()

	go janitor(ctx, lim)

	return lim, nil
}

func (l *Limiter) Allow(key string, cost int64) bool {
	keyHash := hash(key)

	shardIdx := keyHash & (l.Config.shards - 1)

	return l.shards[shardIdx].allow(keyHash, cost, l.Config.supplyRate, l.Config.burst)
}
