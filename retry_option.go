package httpc

type retryOptions struct {
	MaxAttempt      uint
	BackoffStrategy BackoffStrategy
}

var DefaultMaxAttempt uint = 15
var DefaultBackoffStrategy BackoffStrategy = TruncatedExponentialBackoff(10)

type retryOption func(*retryOptions) error

func WithMaxAttempt(maxAttempt uint) retryOption {
	return func(o *retryOptions) error {
		o.MaxAttempt = maxAttempt
		return nil
	}
}

func WithBackoffStrategy(strategy BackoffStrategy) retryOption {
	return func(o *retryOptions) error {
		o.BackoffStrategy = strategy
		return nil
	}
}
