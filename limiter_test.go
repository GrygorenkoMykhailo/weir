package weir

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"
)

// import (
// 	"context"
// 	"strconv"
// 	"sync"
// 	"testing"
// 	"time"
// )

func TestConcurrentAccess(t *testing.T) {
	l, _ := New(context.Background(), RateLimiterOptions{
		Rate:        time.Millisecond,
		Burst:       10,
		Shards:      16,
		KeyTTL:      time.Second,
		CleanupRate: time.Second,
	})

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Go(func() {
			for j := 0; j < 1000; j++ {
				l.Allow(strconv.Itoa(i), 1)
			}
		})
	}

	wg.Wait()
}

func TestNewValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		cfg     RateLimiterOptions
		wantErr error
	}{
		{
			name: "valid config",
			cfg: RateLimiterOptions{
				Rate:        time.Second,
				CleanupRate: time.Second,
				KeyTTL:      time.Second,
				Burst:       10,
				Shards:      256,
			},
			wantErr: nil,
		},
		{
			name: "invalid max tokens",
			cfg: RateLimiterOptions{
				Rate:        time.Second,
				CleanupRate: time.Second,
				KeyTTL:      time.Second,
				Burst:       0,
				Shards:      256,
			},
			wantErr: ErrInvalidBurst,
		},
		{
			name: "invalid cleanup rate",
			cfg: RateLimiterOptions{
				Rate:        time.Second,
				CleanupRate: 0,
				KeyTTL:      time.Second,
				Burst:       10,
				Shards:      100,
			},
			wantErr: ErrInvalidCleanupRate,
		},
		{
			name: "invalid key ttl",
			cfg: RateLimiterOptions{
				Rate:        time.Second,
				CleanupRate: time.Second,
				KeyTTL:      0,
				Burst:       10,
				Shards:      100,
			},
			wantErr: ErrInvalidKeyTTL,
		},
		{
			name: "invalid rate",
			cfg: RateLimiterOptions{
				Rate:        0,
				CleanupRate: time.Second,
				KeyTTL:      time.Second,
				Burst:       10,
				Shards:      100,
			},
			wantErr: ErrInvalidRate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(ctx, tt.cfg)
			if err != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLimiter(t *testing.T) {
	l, err := New(context.Background(), RateLimiterOptions{
		Rate:        100 * time.Millisecond,
		CleanupRate: time.Second,
		KeyTTL:      time.Second,
		Burst:       5,
		Shards:      64,
	})
	if err != nil {
		t.Fatalf("Failed to create limiter: %v", err)
	}

	key := "user-1"

	if !l.Allow(key, 5) {
		t.Fatal("Should allow initial burst of 5")
	}

	if l.Allow(key, 1) {
		t.Fatal("Should deny exceeding limit")
	}

	time.Sleep(110 * time.Millisecond)

	if !l.Allow(key, 1) {
		t.Fatal("Should allow 1 token after refill")
	}
}
