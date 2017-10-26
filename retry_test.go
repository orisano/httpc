package httpc

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetry(t *testing.T) {
	mux := http.NewServeMux()
	s := httptest.NewServer(mux)
	defer s.Close()

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
}
