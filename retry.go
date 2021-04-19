package httpc

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var TimeNow func() time.Time = time.Now
var TimeSleep func(d time.Duration) = time.Sleep

func Retry(client *http.Client, req *http.Request, opts ...RetryOption) (*http.Response, error) {
	if client == nil {
		return nil, fmt.Errorf("missing client")
	}
	if req == nil {
		return nil, fmt.Errorf("missing request")
	}
	options := &retryOptions{
		MaxAttempt:      DefaultMaxAttempt,
		BackoffStrategy: DefaultBackoffStrategy,
	}
	for _, opt := range opts {
		opt(options)
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
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
		attempt++
		if attempt >= options.MaxAttempt {
			return nil, fmt.Errorf("max attempt exceeded")
		}
		if err == nil && len(resp.Header.Get("Retry-After")) > 0 {
			d, err := parseRetryAfter(resp.Header.Get("Retry-After"))
			if err == nil {
				TimeSleep(d)
				continue
			}
		}
		TimeSleep(options.BackoffStrategy.Backoff(attempt))
	}
}

type temporary interface {
	Temporary() bool
}

func isTemporary(err error) bool {
	te, ok := err.(temporary)
	return ok && te.Temporary()
}

type timeout interface {
	Timeout() bool
}

func isTimeout(err error) bool {
	to, ok := err.(timeout)
	return ok && to.Timeout()
}

func isTemporaryStatus(status int) bool {
	switch status {
	case
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusRequestTimeout,
		http.StatusTooManyRequests:
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
	return time.Duration(0), fmt.Errorf("Retry-After header invalid format: %v", ra)
}
