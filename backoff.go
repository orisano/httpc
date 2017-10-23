package httpc

import (
	"math/rand"
	"time"
)

type Backoff interface {
	Backoff(attempt uint) time.Duration
}

type exponentialBackoff struct{}

func ExponentialBackoff() Backoff {
	return &exponentialBackoff{}
}

func (*exponentialBackoff) Backoff(attempt uint) time.Duration {
	return time.Duration(rand.Float64()*(1<<attempt)) * time.Second
}

type truncatedExponentialBackoff struct {
	maxAttempt uint
}

func TruncatedExponentialBackoff(maxAttempt uint) Backoff {
	return &truncatedExponentialBackoff{maxAttempt}
}

func (b *truncatedExponentialBackoff) Backoff(attempt uint) time.Duration {
	n := attempt
	if n > b.maxAttempt {
		n = b.maxAttempt
	}
	return time.Duration(rand.Float64()*(1<<n)) * time.Second
}

type constantBackoff struct {
	duration time.Duration
}

func ConstantBackoff(duration time.Duration) Backoff {
	return &constantBackoff{duration}
}

func (b *constantBackoff) Backoff(attempt uint) time.Duration {
	return b.duration
}
