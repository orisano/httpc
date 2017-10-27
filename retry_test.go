package httpc

import (
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
