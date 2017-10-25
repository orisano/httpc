package httpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestInjectDebugTransport(t *testing.T) {
	send := `{"message": "hello"}`
	response := `{"message": "world"}`

	sendf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer sendf.Close()
	defer os.Remove(sendf.Name())

	fmt.Fprint(sendf, send)
	if err := sendf.Sync(); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if b, err := ioutil.ReadAll(r.Body); err != nil {
			t.Error("request body can't read")
		} else {
			if got := string(b); got != send {
				t.Errorf("unexpected request body. expected: %v, got: %v", send, got)
			}
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, response)
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	f, err := os.Open(sendf.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var b bytes.Buffer
	InjectDebugTransport(http.DefaultClient, &b)
	resp, err := http.Post(s.URL, "application/json", f)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("response body can't read")
	} else {
		if got := string(b); got != response {
			t.Errorf("unexpected response body. expected: %v, got: %v", response, got)
		}
	}
}
