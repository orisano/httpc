package httpc

type retryOptions struct {
	MaxAttempt      uint
	BackoffStrategy BackoffStrategy
}

var DefaultMaxAttempt uint = 15
var DefaultBackoffStrategy BackoffStrategy = TruncatedExponentialBackoff(10)

type retryOption func(*retryOptions)

func WithMaxAttempt(maxAttempt uint) retryOption {
	return func(o *retryOptions) {
		o.MaxAttempt = maxAttempt
	}
}

func WithBackoffStrategy(strategy BackoffStrategy) retryOption {
	return func(o *retryOptions) {
		o.BackoffStrategy = strategy
	}
}
