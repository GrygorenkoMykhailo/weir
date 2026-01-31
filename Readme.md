# Weir

![Go Version](https://img.shields.io/github/go-mod/go-version/GrygorenkoMykhailo/weir)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrygorenkoMykhailo/weir)](https://goreportcard.com/report/github.com/GrygorenkoMykhailo/weir)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

**Weir** is a high-performance, sharded in-memory rate limiter for Go, designed specifically for **high-concurrency** and **high-load** environments.

Unlike standard rate limiters that suffer from mutex contention on multi-core systems, Weir uses a **sharded architecture** to scale linearly with CPU cores.

> **Weir** (noun): A low dam built across a river to regulate the flow of water.


## 🧠 Why Weir?

Most rate limiters (like `x/time/rate`) rely on a **Global Mutex**. This works fine for low traffic but becomes a bottleneck under high concurrency—writes block reads, causing latency spikes during traffic bursts or DDoS attacks.

Weir solves this using **Sharding + Fine-Grained Locking**:

1.  **Sharding:** The key space is divided into thousands of shards (e.g., 1024), effectively spreading lock contention by a factor of 1000.
2.  **Fine-Grained Locking:** Weir separates the "Map Lock" (finding the user) from the "Bucket Lock" (updating the tokens).
    * **Reads are non-blocking:** Checking limits uses parallel `RLock`, allowing thousands of concurrent checks.
    * **Writes are isolated:** Updating a user's token bucket only locks that specific user, not the entire system.


## 🚀 Benchmarks

Running on **AMD Ryzen 7 PRO 5850U (16 logical cores)**.

### Scenario 1: DDoS Attack (Write-Heavy)
*Simulates 1,000,000 distinct keys attacking the system simultaneously. Heavy map churn and locking.*

| Library | ns/op | bytes/op | Speedup |
| :--- | :--- | :--- | :--- |
| **Weir** | **308.6 ns** | **0 B** | **1x** |
| Ulule | 569.9 ns | 63 B | 1.8x slower |
| StdLib | 1673 ns | 9 B | **5.3x slower** |
| Juju | 2009 ns | 11 B | **6.3x slower** |

### Scenario 2: Stable Highload (Read-Heavy)
*Simulates 100,000 active users with a warm cache. Measures standard operation latency.*

| Library | Op/ns | Alloc/op | Speedup |
| :--- | :--- | :--- | :--- |
| **Weir** | **182 ns** | **0 B** | **1x** |
| StdLib | 183 ns | 0 B | ~Same |
| Juju | 184 ns | 0 B | ~Same |
| Ulule | 561 ns | 56 B | 3x slower |

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

	// Check if request is allowed
	if limiter.Allow(userID, 1) {
		fmt.Println("Allowed!")
	} else {
		fmt.Println("Rate limit exceeded (429)")
	}
}
```


## 🌐 Middleware Integrations

Weir is designed to be dependency-free. Integrating Weir is easy—just copy the recipe for your framework.

### Fiber

```go
// Add this middleware to your Fiber app
app.Use(func(c *fiber.Ctx) error {
    // 1. Identify the user (e.g., by IP or API Key)
    key := c.IP()

    // 2. Check the limit (cost = 1)
    if !limiter.Allow(key, 1) {
        // 3. Return 429 if limit exceeded
        return c.Status(fiber.StatusTooManyRequests).SendString("Too Many Requests")
    }

    // 4. Continue
    return c.Next()
})
```

### Gin

```go
// Add this middleware to your Gin router
r.Use(func(c *gin.Context) {
    key := c.ClientIP()

    if !limiter.Allow(key, 1) {
        c.AbortWithStatusJSON(429, gin.H{"error": "Too Many Requests"})
        return
    }

    c.Next()
})
```

### StdLib (`net/http`)

```go
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := r.RemoteAddr // Or extract from headers

        if !limiter.Allow(key, 1) {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Usage:
// http.Handle("/", RateLimitMiddleware(myHandler))
```


## ⚙️ Configuration

| Option | Description | Recommended |
| :--- | :--- | :--- |
| `Rate` | Duration to regenerate **one** token. (e.g. `time.Second / 1000` = 1000 RPS). | `time.Second / RPS` |
| `Burst` | Maximum capacity of the bucket (max burst size). | Based on your load |
| `Shards` | Number of internal map shards. Reduces lock contention. | `512` - `4096` |
| `KeyTTL` | Idle duration after which a key is considered expired. | `1m` - `10m` |
| `CleanupRate` | How often the background janitor scans all shards. | `1m` |


## 📄 License

[MIT](https://choosealicense.com/licenses/mit/)