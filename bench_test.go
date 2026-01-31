package weir

// Before running benchmark download corresponding libraries
//
// go get github.com/juju/ratelimit
// go get github.com/ulule/limiter/v3
// go get golang.org/x/time/rate

// import (
// 	"context"
// 	"fmt"
// 	"math/rand"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/juju/ratelimit"
// 	"github.com/ulule/limiter/v3"
// 	"github.com/ulule/limiter/v3/drivers/store/memory"
// 	"golang.org/x/time/rate"
// )

// type StdLimiter struct {
// 	mu       sync.RWMutex
// 	limiters map[string]*rate.Limiter
// 	r        rate.Limit
// 	b        int
// }

// func NewStdLimiter(r rate.Limit, b int) *StdLimiter {
// 	return &StdLimiter{limiters: make(map[string]*rate.Limiter), r: r, b: b}
// }

// func (l *StdLimiter) Allow(key string) bool {
// 	l.mu.RLock()
// 	lim, exists := l.limiters[key]
// 	l.mu.RUnlock()
// 	if exists {
// 		return lim.Allow()
// 	}
// 	l.mu.Lock()
// 	if lim, exists = l.limiters[key]; exists {
// 		l.mu.Unlock()
// 		return lim.Allow()
// 	}
// 	newLim := rate.NewLimiter(l.r, l.b)
// 	l.limiters[key] = newLim
// 	l.mu.Unlock()
// 	return newLim.Allow()
// }

// type JujuLimiter struct {
// 	mu       sync.RWMutex
// 	buckets  map[string]*ratelimit.Bucket
// 	rate     float64
// 	capacity int64
// }

// func NewJujuLimiter(r float64, c int64) *JujuLimiter {
// 	return &JujuLimiter{buckets: make(map[string]*ratelimit.Bucket), rate: r, capacity: c}
// }

// func (l *JujuLimiter) Allow(key string) bool {
// 	l.mu.RLock()
// 	b, exists := l.buckets[key]
// 	l.mu.RUnlock()
// 	if exists {
// 		return b.TakeAvailable(1) > 0
// 	}
// 	l.mu.Lock()
// 	if b, exists = l.buckets[key]; exists {
// 		l.mu.Unlock()
// 		return b.TakeAvailable(1) > 0
// 	}
// 	newBucket := ratelimit.NewBucketWithRate(l.rate, l.capacity)
// 	l.buckets[key] = newBucket
// 	l.mu.Unlock()
// 	return newBucket.TakeAvailable(1) > 0
// }

// type UluleLimiter struct {
// 	instance *limiter.Limiter
// }

// func NewUluleLimiter(rateStr string) *UluleLimiter {
// 	parsedRate, _ := limiter.NewRateFromFormatted(rateStr)
// 	store := memory.NewStore()
// 	return &UluleLimiter{instance: limiter.New(store, parsedRate)}
// }

// func (l *UluleLimiter) Allow(key string) bool {
// 	ctx, err := l.instance.Get(context.Background(), key)
// 	if err != nil {
// 		return false
// 	}
// 	return !ctx.Reached
// }

// func generateKeys(count int) []string {
// 	keys := make([]string, count)
// 	for i := 0; i < count; i++ {
// 		keys[i] = fmt.Sprintf("user-id-%d-%d", i, rand.Int())
// 	}
// 	return keys
// }

// func Benchmark_Scenario_DDoS_1M(b *testing.B) {
// 	ctx := context.Background()
// 	keysCount := 1_000_000
// 	keys := generateKeys(keysCount)

// 	weirLim, _ := New(ctx, RateLimiterOptions{
// 		Rate: time.Second / 1000, Burst: 100, Shards: 1024, KeyTTL: time.Minute, CleanupRate: time.Minute,
// 	})
// 	stdLim := NewStdLimiter(rate.Limit(1000), 100)
// 	jujuLim := NewJujuLimiter(1000, 100)
// 	ululeLim := NewUluleLimiter("1000-S")

// 	b.ResetTimer()

// 	b.Run("Weir", func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.RunParallel(func(pb *testing.PB) {
// 			var i int
// 			for pb.Next() {
// 				weirLim.Allow(keys[i%keysCount], 1)
// 				i++
// 			}
// 		})
// 	})

// 	b.Run("StdLib", func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.RunParallel(func(pb *testing.PB) {
// 			var i int
// 			for pb.Next() {
// 				stdLim.Allow(keys[i%keysCount])
// 				i++
// 			}
// 		})
// 	})

// 	b.Run("Juju", func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.RunParallel(func(pb *testing.PB) {
// 			var i int
// 			for pb.Next() {
// 				jujuLim.Allow(keys[i%keysCount])
// 				i++
// 			}
// 		})
// 	})

// 	b.Run("Ulule", func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.RunParallel(func(pb *testing.PB) {
// 			var i int
// 			for pb.Next() {
// 				ululeLim.Allow(keys[i%keysCount])
// 				i++
// 			}
// 		})
// 	})
// }

// func Benchmark_Scenario_Stable_100k(b *testing.B) {
// 	ctx := context.Background()
// 	keysCount := 100_000
// 	keys := generateKeys(keysCount)

// 	weirLim, _ := New(ctx, RateLimiterOptions{
// 		Rate: time.Second / 1000, Burst: 100, Shards: 1024, KeyTTL: time.Hour, CleanupRate: time.Hour,
// 	})
// 	stdLim := NewStdLimiter(rate.Limit(1000), 100)
// 	jujuLim := NewJujuLimiter(1000, 100)
// 	ululeLim := NewUluleLimiter("1000-S")

// 	for _, k := range keys {
// 		weirLim.Allow(k, 1)
// 		stdLim.Allow(k)
// 		jujuLim.Allow(k)
// 		ululeLim.Allow(k)
// 	}

// 	b.ResetTimer()

// 	b.Run("Weir", func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.RunParallel(func(pb *testing.PB) {
// 			var i int
// 			for pb.Next() {
// 				weirLim.Allow(keys[i%keysCount], 1)
// 				i++
// 			}
// 		})
// 	})

// 	b.Run("StdLib", func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.RunParallel(func(pb *testing.PB) {
// 			var i int
// 			for pb.Next() {
// 				stdLim.Allow(keys[i%keysCount])
// 				i++
// 			}
// 		})
// 	})

// 	b.Run("Juju", func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.RunParallel(func(pb *testing.PB) {
// 			var i int
// 			for pb.Next() {
// 				jujuLim.Allow(keys[i%keysCount])
// 				i++
// 			}
// 		})
// 	})

// 	b.Run("Ulule", func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.RunParallel(func(pb *testing.PB) {
// 			var i int
// 			for pb.Next() {
// 				ululeLim.Allow(keys[i%keysCount])
// 				i++
// 			}
// 		})
// 	})
// }
