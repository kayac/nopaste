package nopaste

import (
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	last := time.Now()
	start := time.Now()
	for i := 0; i < 10; i++ {
		throttle(last, 100*time.Millisecond)
		last = time.Now()
	}
	elapsed := last.Sub(start)
	if elapsed.Seconds() < 1 {
		t.Errorf("elapsed %s < 1s", elapsed)
	}
	if elapsed.Seconds() > 1.5 {
		t.Errorf("elapsed %s > 1.5s", elapsed)
	}
}
