package httpc

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

var TimeNow func() time.Time = time.Now

func Retry(client *http.Client, req *http.Request, opts ...retryOption) (*http.Response, error) {
	if client == nil {
		return nil, errors.New("missing client")
	}
	if req == nil {
		return nil, errors.New("missing req")
	}
	options := &retryOptions{
		MaxAttempt:      DefaultMaxAttempt,
		BackoffStrategy: DefaultBackoffStrategy,
	}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, errors.Wrap(err, "failed to apply option")
		}
	}

	attempt := uint(0)
	for {
		req.Close = false
		resp, err := client.Do(req)
		if err != nil {
			if !isTimeout(err) && !isTemporary(err) {
				return nil, err
			}
		} else {
			if !isTemporaryStatus(resp.StatusCode) {
				return resp, nil
			}
			resp.Body.Close()
		}
		attempt++
		if attempt >= options.MaxAttempt {
			return nil, errors.New("max attempt exceeded")
		}
		if err == nil && len(resp.Header.Get("Retry-After")) > 0 {
			d, err := parseRetryAfter(resp.Header.Get("Retry-After"))
			if err == nil {
				time.Sleep(d)
				continue
			}
		}
		time.Sleep(options.BackoffStrategy.Backoff(attempt))
	}
}

type temporary interface {
	Temporary() bool
}

func isTemporary(err error) bool {
	te, ok := errors.Cause(err).(temporary)
	return ok && te.Temporary()
}

type timeout interface {
	Timeout() bool
}

func isTimeout(err error) bool {
	to, ok := errors.Cause(err).(timeout)
	return ok && to.Timeout()
}

func isTemporaryStatus(status int) bool {
	switch status {
	case http.StatusInternalServerError:
	case http.StatusBadGateway:
	case http.StatusServiceUnavailable:
	case http.StatusGatewayTimeout:
	case http.StatusRequestTimeout:
	case http.StatusTooManyRequests:
		return true
	}
	return false
}

func parseRetryAfter(ra string) (time.Duration, error) {
	if d, err := http.ParseTime(ra); err == nil {
		return d.Sub(TimeNow()), nil
	}
	if s, err := strconv.ParseUint(ra, 10, 32); err == nil {
		return time.Duration(s) * time.Second, nil
	}
	return time.Duration(0), errors.Errorf("Retry-After header invalid format: %v", ra)
}