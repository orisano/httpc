package httpc

import (
	"math/rand"
	"time"
)

type BackoffStrategy interface {
	Backoff(attempt uint) time.Duration
}

type exponentialBackoff struct{}

func ExponentialBackoff() BackoffStrategy {
	return &exponentialBackoff{}
}

func (*exponentialBackoff) Backoff(attempt uint) time.Duration {
	return time.Duration(rand.Float64()*(1<<attempt)) * time.Second
}

type truncatedExponentialBackoff struct {
	maxAttempt uint
}

func TruncatedExponentialBackoff(maxAttempt uint) BackoffStrategy {
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

func ConstantBackoff(duration time.Duration) BackoffStrategy {
	return &constantBackoff{duration}
}

func (b *constantBackoff) Backoff(attempt uint) time.Duration {
	return b.duration
}
