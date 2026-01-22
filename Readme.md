# Weir

![Go Version](https://img.shields.io/github/go-mod/go-version/GrygorenkoMykhailo/weir)
![Go Report Card](https://goreportcard.com/badge/github.com/GrygorenkoMykhailo/weir)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

**Weir** is a high-performance, sharded in-memory rate limiter for Go, designed specifically for **high-concurrency** and **high-load** environments.

Unlike standard rate limiters that suffer from mutex contention on multi-core systems, Weir uses a **sharded architecture** to scale linearly with CPU cores.

> **Weir** (noun): A low dam built across a river to regulate the flow of water.

## 🚀 Benchmarks

Running on **AMD Ryzen 7 PRO 5850U (16 logical cores)**.

Weir is **~13x faster** than the standard library (`golang.org/x/time/rate`) and **~30x faster** than `juju/ratelimit` under heavy write load (DDOS scenario).

| Library | Scenario | Op/ns | Alloc/op | Speedup |
| :--- | :--- | :--- | :--- | :--- |
| **Weir** | **DDOS (Write-Heavy)** | **193 ns** | **106 B** | **1x (Baseline)** |
| Ulule | DDOS (Write-Heavy) | 397 ns | 228 B | 2x slower |
| StdLib | DDOS (Write-Heavy) | 2541 ns | 207 B | **13x slower** |
| Juju | DDOS (Write-Heavy) | 5984 ns | 186 B | **31x slower** |
| | | | | |
| **Weir** | **Stable (Read-Heavy)** | **185 ns** | **0 B** | **1x (Baseline)** |
| Ulule | Stable (Read-Heavy) | 536 ns | 39 B | 2.9x slower |
| StdLib | Stable (Read-Heavy) | 181 ns | 0 B | ~Same |
| Juju | Stable (Read-Heavy) | 182 ns | 186 B | ~Same |

## 📦 Installation

```bash
go get [github.com/GrygorenkoMykhailo/weir](https://github.com/GrygorenkoMykhailo/weir)
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

	"[github.com/GrygorenkoMykhailo/weir](https://github.com/GrygorenkoMykhailo/weir)"
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
    "[github.com/gofiber/fiber/v2](https://github.com/gofiber/fiber/v2)"
    // Alias used to avoid conflict with "fiber" package
    weirmw "[github.com/GrygorenkoMykhailo/weir/middlewares/fiber](https://github.com/GrygorenkoMykhailo/weir/middlewares/fiber)"
)

// ...

app.Use(weirmw.Middleware(1, func(c *fiber.Ctx) error {
    return c.Next()
}, &weirmw.FiberMiddlewareOptions{
    Limiter: myLimiter,
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
    "[github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)"
    // Alias used to avoid conflict with "gin" package
    weirmw "[github.com/GrygorenkoMykhailo/weir/middlewares/gin](https://github.com/GrygorenkoMykhailo/weir/middlewares/gin)"
)

// ...

r.Use(weirmw.Middleware(1, func(c *gin.Context) {
    c.Next()
}, &weirmw.GinMiddlewareOptions{
    Limiter: myLimiter,
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
    "net/http"
    "[github.com/GrygorenkoMykhailo/weir/middlewares/stdlib](https://github.com/GrygorenkoMykhailo/weir/middlewares/stdlib)"
)

// ...

http.HandleFunc("/", stdlib.Middleware(1, myHandler, &stdlib.StdLibMiddlewareOptions{
    Limiter: myLimiter,
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
| `Shards` | Number of internal map shards. Reduces lock contention. | `512` - `4096` |
| `KeyTTL` | Idle duration after which a user key is removed. | `1m` - `10m` |
| `CleanupRate` | How often the background janitor scans all shards. | `1m` |

## 🧠 Why is it so fast?

### The Problem: Global Mutex Contention
Standard libraries typically use a single `sync.RWMutex` to protect the map of limiters. When 16+ cores try to write to this map simultaneously (e.g., during a DDOS attack), the CPU spends most of its time waiting for locks rather than doing work.

### The Solution: Sharding
Weir partitions the key space into **Shards**.
1.  **Low Contention:** Each request locks only one shard of the memory. Threads rarely block each other.
2.  **Probabilistic Janitor:** A background "janitor" cleans up expired keys using a Redis-style probabilistic algorithm, ensuring O(1) blocking time regardless of map size.

## 📄 License

[MIT](https://choosealicense.com/licenses/mit/)