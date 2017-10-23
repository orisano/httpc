package httpc

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

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
		resp, err := client.Do(req)
		if err != nil {
			if !isTimeout(err) && !isTemporary(err) {
				return nil, err
			}
		} else {

		}
		if err == nil {
			resp.Body.Close()

		}
		attempt++
		if attempt >= options.MaxAttempt {
			return nil, errors.New("max attempt exceeded")
		}

	}
}

func isAutoReuse(r io.Reader) bool {
	switch r.(type) {
	case *bytes.Buffer:
	case *bytes.Reader:
	case *strings.Reader:
		return true
	}
	return false
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
		return true
	}
	return false
}
