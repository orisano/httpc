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
	t.Run("PositiveCase", func(t *testing.T) {
		send := `{"message": "hello"}`
		response := `{"message": "world"}`

		tf, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tf.Name())
		defer tf.Close()

		tf.WriteString(send)
		tf.Sync()

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

		f, err := os.Open(tf.Name())
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
		if b.Len() == 0 {
			t.Error("dump buffer size == 0")
		}
	})
	t.Run("NilClient", func(t *testing.T) {
		var b bytes.Buffer
		err := InjectDebugTransport(nil, &b)
		if err == nil {
			t.Errorf("accept nil client")
		}
	})
	t.Run("NilWriter", func(t *testing.T) {
		err := InjectDebugTransport(http.DefaultClient, nil)
		if err == nil {
			t.Errorf("accept nil writer")
		}
	})
}

func TestRemoveDebugTransport(t *testing.T) {
	t.Run("NilClient", func(t *testing.T) {
		err := RemoveDebugTransport(nil)
		if err == nil {
			t.Errorf("accept nil client")
		}
	})
	t.Run("NoInjectedClient", func(t *testing.T) {
		c := &http.Client{}
		if err := RemoveDebugTransport(c); err != nil {
			t.Error(err)
		}
	})
	t.Run("PositiveCase", func(t *testing.T) {
		rt := http.DefaultTransport
		c := &http.Client{Transport: rt}
		b := new(bytes.Buffer)

		if err := InjectDebugTransport(c, b); err != nil {
			t.Fatal(err)
		}
		if err := RemoveDebugTransport(c); err != nil {
			t.Fatal(err)
		}
		if got := c.Transport; got != rt {
			t.Errorf("unexpected transport. expected: %v, but got: %v", rt, got)
		}
	})
}
