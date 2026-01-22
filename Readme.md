# Weir

![Go Version](https://img.shields.io/github/go-mod/go-version/GrygorenkoMykhailo/weir)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrygorenkoMykhailo/weir)](https://goreportcard.com/report/github.com/GrygorenkoMykhailo/weir)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

**Weir** is a high-performance, sharded in-memory rate limiter for Go, designed specifically for **high-concurrency** and **high-load** environments.

Unlike standard rate limiters that suffer from mutex contention on multi-core systems, Weir uses a **sharded architecture** to scale linearly with CPU cores.

> **Weir** (noun): A low dam built across a river to regulate the flow of water.


## 🧠 Why Weir? (The Engineering Trade-off)

Weir is designed for **system stability**, not just raw micro-benchmark speed.

### The Problem with Standard Limiters
Standard limiters (like `x/time/rate` wrapped in a map) use a **Global Mutex**.
* **Happy Path (Reads):** They are fast because RWMutex is optimized for reads.
* **Highload (Writes/Churn):** When new users arrive (traffic spikes, DDOS, cache rotation), the Global Write Lock **blocks everyone**. Performance degrades significantly.

### The Weir Solution
Weir uses **Sharding**.
* We pay a tiny "tax" for hashing and shard selection on every request.
* **In return, we get 100% predictable latency.**
* Write heavy loads (DDOS) do not block existing users.
* Memory allocations are absent when reading and minimum **2x** smaller when writing.
* Background janitor cleans up expired keys using probabilistic algoritm, ensuring tiniest possible locking times
regardless of keys amount


## 🚀 Benchmarks

Running on **AMD Ryzen 7 PRO 5850U (16 logical cores)**.

Weir is **~13x faster** than the standard library (`golang.org/x/time/rate`), **~2x faster** than (`github.com/ulule/limiter`) and **~31x faster** than (`github.com/juju/ratelimit`) under heavy write load (Highload scenario).

| Library | Scenario | Op/ns | Alloc/op | Speedup |
| :--- | :--- | :--- | :--- | :--- |
| **Weir** | **DDOS (Write-Heavy)** | **193 ns** | **106 B** | **1x (Baseline)** |
| Ulule | Highload (Write-Heavy) | 397 ns | 228 B | 2x slower |
| StdLib | Highload (Write-Heavy) | 2541 ns | 207 B | **13x slower** |
| Juju | Highload (Write-Heavy) | 5984 ns | 186 B | **31x slower** |
| | | | | |
| **Weir** | **Stable (Read-Heavy)** | **185 ns** | **0 B** | **1x (Baseline)** |
| Ulule | Stable (Read-Heavy) | 536 ns | 39 B | 2.9x slower |
| StdLib | Stable (Read-Heavy) | 181 ns | 0 B | ~Same |
| Juju | Stable (Read-Heavy) | 182 ns | 0 B | ~Same |


## 📦 Installation

```bash
go get github.com/GrygorenkoMykhailo/weir
```

## ⚡ Quick Start

### Core Library

Weir provides a simple, idiomatic API.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/GrygorenkoMykhailo/weir"
)

func main() {
	ctx := context.Background()

	// Initialize the limiter
	// Example: 1000 requests/sec, Burst 50
	limiter, err := weir.New(ctx, weir.RateLimiterOptions{
		Rate:        time.Second / 1000, // 1 token every 1ms = 1000 RPS
		Burst:       50,                 // Bucket capacity
		Shards:      4096,               // Shard count (power of 2)
		KeyTTL:      time.Minute,        // Clean up idle keys after 1m
		CleanupRate: time.Minute,        // Run janitor every 1m
	})
	if err != nil {
		panic(err)
	}

	userID := "user-123"
	cost := int64(1)

	// Check if request is allowed
	if limiter.Allow(userID, cost) {
		fmt.Println("Allowed!")
	} else {
		fmt.Println("Rate limit exceeded (429)")
	}
}
```


## 🌐 Middleware Integrations

Weir is easy to integrate with popular Go web frameworks.

### Fiber

```go
import (
    "time"
    "context"
    "github.com/gofiber/fiber/v2"
    "github.com/GrygorenkoMykhailo/weir"
    // Alias used to avoid conflict with "fiber" package
    weirmw "github.com/GrygorenkoMykhailo/weir/middlewares/fiber"
)

// ...

// Create Limiter Instance
limiter, _ := weir.New(context.Background(), weir.RateLimiterOptions{
    Rate: time.Second / 100, Burst: 20, Shards: 1024,
    KeyTTL: time.Minute, CleanupRate: time.Minute,
})

app.Use(weirmw.Middleware(1, func(c *fiber.Ctx) error {
    return c.Next()
}, &weirmw.FiberMiddlewareOptions{
    Limiter: limiter,
    KeyExtractor: func(c *fiber.Ctx) string {
        return c.IP()
    },
    OnTooManyRequests: func(c *fiber.Ctx) error {
        return c.Status(429).SendString("Too Many Requests")
    },
}))
```

### Gin

```go
import (
    "time"
    "context"
    "github.com/gin-gonic/gin"
    "github.com/GrygorenkoMykhailo/weir"
    // Alias used to avoid conflict with "gin" package
    weirmw "github.com/GrygorenkoMykhailo/weir/middlewares/gin"
)

// ...

// Create Limiter Instance
limiter, _ := weir.New(context.Background(), weir.RateLimiterOptions{
    Rate: time.Second / 100, Burst: 20, Shards: 1024,
    KeyTTL: time.Minute, CleanupRate: time.Minute,
})

r.Use(weirmw.Middleware(1, func(c *gin.Context) {
    c.Next()
}, &weirmw.GinMiddlewareOptions{
    Limiter: limiter,
    KeyExtractor: func(c *gin.Context) string {
        return c.ClientIP()
    },
    OnTooManyRequests: func(c *gin.Context) {
        c.AbortWithStatus(429)
    },
}))
```

### StdLib (`net/http`)

```go
import (
    "time"
    "context"
    "net/http"
    "github.com/GrygorenkoMykhailo/weir"
    "github.com/GrygorenkoMykhailo/weir/middlewares/stdlib"
)

// ...

// Create Limiter Instance
limiter, _ := weir.New(context.Background(), weir.RateLimiterOptions{
    Rate: time.Second / 100, Burst: 20, Shards: 1024,
    KeyTTL: time.Minute, CleanupRate: time.Minute,
})

http.HandleFunc("/", stdlib.Middleware(1, myHandler, &stdlib.StdLibMiddlewareOptions{
    Limiter: limiter,
    KeyExtractor: func(r *http.Request) string {
        return r.RemoteAddr
    },
    OnTooManyRequests: func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(429)
    },
}))
```


## ⚙️ Configuration

| Option | Description | Recommended |
| :--- | :--- | :--- |
| `Rate` | Duration to regenerate **one** token. (e.g. `time.Second / 1000` = 1000 RPS). | `time.Second / RPS` |
| `Burst` | Maximum capacity of the bucket (max burst size). | Based on your load |
| `Shards` | Number of internal map shards. Reduces lock contention. **Should be a power of two**. | `512` - `4096` |
| `KeyTTL` | Idle duration after which a user key is removed. | `1m` - `10m` |
| `CleanupRate` | How often the background janitor scans all shards. | `1m` |


## 📄 License

[MIT](https://choosealicense.com/licenses/mit/)