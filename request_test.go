package httpc

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

func readAllString(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func TestNewRequest(t *testing.T) {
	rawurl := "http://web.example"

	t.Run("WithBody", func(t *testing.T) {
		expected := "Test Request"

		body := bytes.NewReader([]byte(expected))
		req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
			WithBody(body),
		)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := readAllString(req.Body); got != expected {
			t.Errorf("unexpected request body. expected: %v, got: %v", expected, got)
		} else if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("WithBinary", func(t *testing.T) {
		expected := "Test Binary"

		body := bytes.NewReader([]byte(expected))
		req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
			WithBinary(body),
		)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := readAllString(req.Body); got != expected {
			t.Errorf("unexpected request body. expected: %v, got: %v", expected, got)
		} else if err != nil {
			t.Fatal(err)
		}
		if got := req.Header.Get("Content-Type"); got != "application/octet-stream" {
			t.Errorf("unexpected content-type. expected: application/octet-stream, got: %v", got)
		}
	})

	t.Run("WithForm", func(t *testing.T) {
		expected := url.Values{}
		expected.Set("id", "john")
		expected.Set("password", "dummy_password")

		req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
			WithForm(expected),
		)
		if err != nil {
			t.Fatal(err)
		}
		if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected content-type. expected: application/x-www-form-urlencoded, got: %v", got)
		}

		req.ParseForm()
		for _, key := range []string{"id", "password"} {
			if got := req.Form.Get(key); got != expected.Get(key) {
				t.Errorf("unexpected %v. expected: %v, got: %v", key, expected.Get(key), got)
			}
		}
	})

	t.Run("WithJSON", func(t *testing.T) {
		type testStruct struct {
			Icon string `json:"icon"`
			Text string `json:"text"`
		}
		expected := testStruct{
			Icon: "http://web.example/icons/icon.png",
			Text: "hello from test bot",
		}
		expectedJSON := `{"icon":"http://web.example/icons/icon.png","text":"hello from test bot"}` + "\n"

		req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
			WithJSON(expected),
		)
		if err != nil {
			t.Fatal(err)
		}
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("unexpected content-type. expected: application/json, got: %v", got)
		}
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(b); got != expectedJSON {
			t.Errorf("unexpected json body. expected: %v, got: %v", expectedJSON, got)
		}
		var got testStruct
		if err := json.NewDecoder(bytes.NewReader(b)).Decode(&got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected data. expected: %+v, got: %+v", expected, got)
		}
	})

	t.Run("WithXML", func(t *testing.T) {
		type User struct {
			XMLName xml.Name
			Name    string
			Email   string
			Age     int
			Weight  float64
		}
		expected := User{
			Name:   "admin",
			Email:  "webmaster@mail.example",
			Age:    17,
			Weight: 45.1,
		}
		expectedXML := xml.Header + `<User><Name>admin</Name><Email>webmaster@mail.example</Email><Age>17</Age><Weight>45.1</Weight></User>`

		req, err := NewRequest(context.TODO(), http.MethodPut, rawurl,
			WithXML(expected),
		)
		if err != nil {
			t.Fatal(err)
		}
		if got := req.Header.Get("Content-Type"); got != `application/xml; charset="UTF-8"` {
			t.Errorf(`unexpected content-type. expected: application/xml; charset="UTF-8", got: %v`, got)
		}
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(b); got != expectedXML {
			t.Errorf("unexpected xml body. expected: %v, got: %v", expectedXML, got)
		}
		var got User
		if err := xml.NewDecoder(bytes.NewReader(b)).Decode(&got); err != nil {
			t.Fatal(err)
		}
		got.XMLName.Local = ""
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected data. expected: %+v, got: %+v", expected, got)
		}
	})

	t.Run("AddHeaderField", func(t *testing.T) {
		expected := "orisano-httpc/1.0"
		req, err := NewRequest(context.TODO(), http.MethodGet, rawurl,
			AddHeaderField("User-Agent", expected),
		)
		if err != nil {
			t.Fatal(err)
		}
		if got := req.UserAgent(); got != expected {
			t.Errorf("unexpected user-agent. expected: %v, got: %v", expected, got)
		}
	})

	t.Run("WithHeader", func(t *testing.T) {
		expected := http.Header{}
		expected.Add("Authentication", "Bearer xxxxxxxxxxxx-xxxxxxxxxxx-xxxxxxxxxxxx")
		expected.Add("X-API-Version", "2017-10-26")

		req, err := NewRequest(context.TODO(), http.MethodGet, rawurl,
			WithHeader(expected),
		)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"Authentication", "X-API-Version"} {
			if got := req.Header.Get(key); got != expected.Get(key) {
				t.Errorf("unexpected %v. expected: %v, got: %v", key, expected.Get(key), got)
			}
		}
	})

	t.Run("AddQuery", func(t *testing.T) {
		expected := "3"
		req, err := NewRequest(context.TODO(), http.MethodGet, rawurl,
			AddQuery("page", expected),
		)
		if err != nil {
			t.Fatal(err)
		}
		if got := req.URL.Query().Get("page"); got != expected {
			t.Errorf("unexpected page param. expected: %v, got: %v", expected, got)
		}
	})

	t.Run("WithQueries", func(t *testing.T) {
		expected := url.Values{}
		expected.Set("utf8", "✓")
		expected.Set("q", "Error")

		req, err := NewRequest(context.TODO(), http.MethodGet, rawurl,
			WithQueries(expected),
		)
		if err != nil {
			t.Fatal(err)
		}

		q := req.URL.Query()
		for _, key := range []string{"utf8", "q"} {
			if got := q.Get(key); got != expected.Get(key) {
				t.Errorf("unexpected %v param. expected: %v, got: %v", key, expected.Get(key), got)
			}
		}
	})

	t.Run("EnforceContentLength", func(t *testing.T) {
		tf, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tf.Name())
		defer tf.Close()
		tf.WriteString("ContentLength")
		tf.Sync()

		t.Run("NotEnforce", func(t *testing.T) {
			expected := int64(0)
			f, err := os.Open(tf.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
				WithBody(f),
			)
			if err != nil {
				t.Fatal(err)
			}
			if got := req.ContentLength; got != expected {
				t.Errorf("unexpected ContentLength. expected: %v, got: %v", expected, got)
			}
		})

		t.Run("Enforce", func(t *testing.T) {
			expected := int64(len("ContentLength"))
			f, err := os.Open(tf.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
				WithBody(f),
				EnforceContentLength,
			)
			if err != nil {
				t.Fatal(err)
			}
			if got := req.ContentLength; got != expected {
				t.Errorf("unexpected ContentLength. expected: %v, got: %v", expected, got)
			}
		})

		// ContentLength automatically set structs. strings.Reader, bytes.Reader, bytes.Buffer
		t.Run("Automatically", func(t *testing.T) {
			s := "Automatically"
			expected := int64(len(s))

			b := strings.NewReader(s)
			req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
				WithBody(b),
			)
			if err != nil {
				t.Fatal(err)
			}
			if got := req.ContentLength; got != expected {
				t.Errorf("unexpected ContentLength. expected: %v, got: %v", expected, got)
			}
		})
	})

	t.Run("OverridePattern", func(t *testing.T) {
		t.Run("Query", func(t *testing.T) {
			expected := ""
			queries := url.Values{}
			queries.Set("a", "1")
			queries.Set("b", "2")

			req, err := NewRequest(context.TODO(), http.MethodGet, rawurl,
				AddQuery("c", "3"),
				WithQueries(queries),
			)
			if err != nil {
				t.Fatal(err)
			}
			if got := req.URL.Query().Get("c"); got != expected {
				t.Errorf("unexpected c param. expected: %v, got: %v", expected, got)
			}
		})
		t.Run("Header", func(t *testing.T) {
			expected := ""
			header := http.Header{}
			header.Set("a", "1")
			header.Set("b", "2")

			req, err := NewRequest(context.TODO(), http.MethodGet, rawurl,
				AddHeaderField("c", "3"),
				WithHeader(header),
			)
			if err != nil {
				t.Fatal(err)
			}
			if got := req.Header.Get("c"); got != expected {
				t.Errorf("unexpected c header. expected: %v, got: %v", expected, got)
			}
		})
	})

	t.Run("NilContext", func(t *testing.T) {
		req, err := NewRequest(nil, http.MethodGet, rawurl)
		if err == nil {
			t.Errorf("accept nil context")
		}
		if req != nil {
			t.Errorf("return invalid request")
		}
	})

	t.Run("EmptyMethod", func(t *testing.T) {
		req, err := NewRequest(context.TODO(), "", rawurl)
		if err == nil {
			t.Errorf("accept empty method")
		}
		if req != nil {
			t.Errorf("return invalid request")
		}
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		req, err := NewRequest(context.TODO(), " ", rawurl)
		if err == nil {
			t.Error("accept invalid method")
		}
		if req != nil {
			t.Errorf("return invalid request")
		}
	})

	t.Run("EmptyURL", func(t *testing.T) {
		req, err := NewRequest(context.TODO(), http.MethodConnect, "")
		if err == nil {
			t.Errorf("accept empty url")
		}
		if req != nil {
			t.Errorf("return invalid request")
		}
	})

	t.Run("InvalidURL", func(t *testing.T) {
		req, err := NewRequest(context.TODO(), http.MethodDelete, "web.invalid",
			AddQuery("parse", "fire"),
		)
		if err == nil {
			t.Errorf("accept invalid url")
		}
		if req != nil {
			t.Errorf("return invalid request")
		}
	})

	t.Run("UnreadableBody", func(t *testing.T) {
		req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
			WithBody(&unreadable{}),
			EnforceContentLength,
		)
		if err == nil {
			t.Errorf("read unreadable body")
		}
		if req != nil {
			t.Errorf("return invalid request")
		}
	})

	t.Run("CantMarshalJSON", func(t *testing.T) {
		req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
			WithJSON(&cantMarshalJSON{}),
		)
		if err == nil {
			t.Errorf("accept cant MarshalJSON")
		}
		if req != nil {
			t.Errorf("return invalid request")
		}
	})

	t.Run("CantMarshalXML", func(t *testing.T) {
		req, err := NewRequest(context.TODO(), http.MethodPost, rawurl,
			WithXML(&cantMarshalXML{}),
		)
		if err == nil {
			t.Errorf("accept cant MarshalXML")
		}
		if req != nil {
			t.Errorf("return invalid request")
		}
	})
}

type unreadable struct{}

func (*unreadable) Read(p []byte) (n int, err error) {
	return 0, errors.New("read failed")
}

type cantMarshalJSON struct{}

func (*cantMarshalJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("cant")
}

type cantMarshalXML struct{}

func (*cantMarshalXML) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return errors.New("cant")
}
