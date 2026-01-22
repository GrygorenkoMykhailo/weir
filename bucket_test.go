package weir

import (
	"testing"
	"time"
)

func TestBucketRefill(t *testing.T) {
	rate := (100 * time.Millisecond).Nanoseconds()
	Burst := int64(10)

	now := int64(0)

	b := &bucket{
		tokens:     Burst,
		lastCalled: now,
	}

	if !b.allow(10, rate, Burst, now) {
		t.Fatal("step 1: should have allowed taking all 10 tokens initially")
	}

	if b.tokens != 0 {
		t.Fatalf("step 1: expected 0 tokens left, got %d", b.tokens)
	}

	if b.allow(1, rate, Burst, now) {
		t.Fatal("step 1: bucket should be empty immediately after draining")
	}

	now += (300 * time.Millisecond).Nanoseconds()

	if !b.allow(3, rate, Burst, now) {
		t.Errorf("step 2: should allow taking 3 tokens after 300ms refill. Tokens available: %d", b.tokens)
	}

	if b.allow(1, rate, Burst, now) {
		t.Fatal("step 2: bucket should be empty again")
	}

	now += time.Hour.Nanoseconds()

	if !b.allow(10, rate, Burst, now) {
		t.Fatal("step 3: should allow full burst after long wait")
	}

	if b.allow(1, rate, Burst, now) {
		t.Fatal("step 3: tokens should not exceed Burst limit (cap failed)")
	}
}
