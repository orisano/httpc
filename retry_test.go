package httpc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	TimeSleep = func(d time.Duration) {}

	t.Run("NilClient", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, s.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := Retry(nil, req)
		if err == nil {
			t.Errorf("accept nil client")
		}
		if resp != nil {
			t.Errorf("return invalid response")
		}
	})
	t.Run("NilRequest", func(t *testing.T) {
		resp, err := Retry(http.DefaultClient, nil)
		if err == nil {
			t.Errorf("accept nil request")
		}
		if resp != nil {
			t.Errorf("return invalid response")
		}
	})
	t.Run("SuccessRequest", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, s.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := Retry(http.DefaultClient, req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
	})
	t.Run("Retry", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, s.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		ts := []struct {
			Transport  http.RoundTripper
			MaxAttempt uint
			IsError    bool
		}{
			{
				Transport:  &errorTransport{fmt.Errorf("normal"), 1},
				MaxAttempt: 10,
				IsError:    true,
			},
			{
				Transport:  &errorTransport{&timeoutError{}, 4},
				MaxAttempt: 5,
				IsError:    false,
			},
			{
				Transport:  &errorTransport{&timeoutError{}, 5},
				MaxAttempt: 5,
				IsError:    true,
			},
			{
				Transport:  &errorTransport{&temporaryError{}, 9},
				MaxAttempt: 10,
				IsError:    false,
			},
			{
				Transport:  &errorTransport{&temporaryError{}, 10},
				MaxAttempt: 10,
				IsError:    true,
			},
			{
				Transport:  &statusTransport{http.StatusInternalServerError, 10},
				MaxAttempt: 15,
				IsError:    false,
			},
			{
				Transport:  &statusTransport{http.StatusBadGateway, 20},
				MaxAttempt: 25,
				IsError:    false,
			},
			{
				Transport:  &statusTransport{http.StatusRequestTimeout, 10},
				MaxAttempt: 15,
				IsError:    false,
			},
			{
				Transport:  &statusTransport{http.StatusTooManyRequests, 10},
				MaxAttempt: 15,
				IsError:    false,
			},
		}

		for _, tc := range ts {
			client := &http.Client{
				Transport: tc.Transport,
			}
			resp, err := Retry(client, req, WithMaxAttempt(tc.MaxAttempt))
			if tc.IsError {
				if err == nil {
					t.Errorf("request must be fail")
				}
			} else {
				if err != nil {
					t.Errorf("request must be success")
				}
				if got := resp.StatusCode; got != http.StatusOK {
					t.Errorf("unexpected status code. expected: %v, got: %v", http.StatusOK, got)
				}
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	})
	t.Run("RetryAfter", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, s.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		for _, tc := range []string{"3000", "Wed, 21 Oct 2015 07:28:00 GMT"} {
			resp, err := Retry(
				&http.Client{
					Transport: &retryAfterTransport{tc, 1},
				},
				req,
				WithMaxAttempt(2),
				WithBackoffStrategy(&panicBackoff{}),
			)
			if err != nil {
				t.Fatal(err)
			}
			if got := resp.StatusCode; got != http.StatusOK {
				t.Errorf("unexpected status code. expected: %v, got: %v", resp.StatusCode, got)
			}
			resp.Body.Close()
		}
	})
}

type timeoutError struct{}

func (*timeoutError) Timeout() bool {
	return true
}

func (*timeoutError) Error() string {
	return "timeout"
}

type errorTransport struct {
	err   error
	count int
}

func (t *errorTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.count == 0 {
		return http.DefaultTransport.RoundTrip(r)
	} else {
		t.count--
		return nil, t.err
	}
}

type temporaryError struct{}

func (*temporaryError) Error() string {
	return "temporary"
}

func (*temporaryError) Temporary() bool {
	return true
}

type statusTransport struct {
	code  int
	count int
}

func (t *statusTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if t.count > 0 {
		if resp.StatusCode == http.StatusOK {
			resp.StatusCode = t.code
		}
		t.count--
	}
	return resp, nil
}

type panicBackoff struct{}

func (*panicBackoff) Backoff(attempt uint) time.Duration {
	panic("panic")
}

type retryAfterTransport struct {
	ra    string
	count int
}

func (t *retryAfterTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if t.count > 0 {
		t.count--
		resp.StatusCode = http.StatusTooManyRequests
		resp.Header.Set("Retry-After", t.ra)
	}
	return resp, nil
}
