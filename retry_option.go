package httpc

type retryOptions struct {
	MaxAttempt      uint
	BackoffStrategy BackoffStrategy
}

var DefaultMaxAttempt uint = 15
var DefaultBackoffStrategy BackoffStrategy = TruncatedExponentialBackoff(6)

type RetryOption func(*retryOptions)

func WithMaxAttempt(maxAttempt uint) RetryOption {
	return func(o *retryOptions) {
		o.MaxAttempt = maxAttempt
	}
}

func WithBackoffStrategy(strategy BackoffStrategy) RetryOption {
	return func(o *retryOptions) {
		o.BackoffStrategy = strategy
	}
}
