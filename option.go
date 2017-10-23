package httpc

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Options struct {
	Body   io.Reader
	Header http.Header
	Params url.Values

	EnforceContentLength bool
}

type Option func(*Options) error

func ComposeOption(opts ...Option) Option {
	return func(o *Options) error {
		for _, opt := range opts {
			if err := opt(o); err != nil {
				return err
			}
		}
		return nil
	}
}

func ApplyOption(o *Options, opts ...Option) error {
	return ComposeOption(opts...)(o)
}

func WithBody(body io.Reader) Option {
	return func(o *Options) error {
		o.Body = body
		return nil
	}
}

func WithBinary(bin io.Reader) Option {
	return func(o *Options) error {
		o.Header.Add("Content-Type", "application/octet-stream")
		o.Body = bin
		return nil
	}
}

func WithForm(params url.Values) Option {
	return func(o *Options) error {
		o.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		o.Body = strings.NewReader(params.Encode())
		return nil
	}
}

func WithJSON(data interface{}) Option {
	return func(o *Options) error {
		o.Header.Set("Content-Type", "application/json")
		r, w := io.Pipe()
		go func() {
			json.NewEncoder(w).Encode(data)
			w.Close()
		}()
		o.Body = r
		return nil
	}
}

func WithXML(data interface{}) Option {
	return func(o *Options) error {
		o.Header.Set("Content-Type", "application/xml; charset=\"UTF-8\"")
		r, w := io.Pipe()
		go func() {
			w.Write([]byte(xml.Header))
			xml.NewEncoder(w).Encode(data)
			w.Close()
		}()
		o.Body = r
		return nil
	}
}

func AddHeader(key, value string) Option {
	return func(o *Options) error {
		o.Header.Add(key, value)
		return nil
	}
}

func WithHeader(header http.Header) Option {
	return func(o *Options) error {
		o.Header = header
		return nil
	}
}

func AddParams(key, value string) Option {
	return func(o *Options) error {
		o.Params.Add(key, value)
		return nil
	}
}

func WithParams(params url.Values) Option {
	return func(o *Options) error {
		o.Params = params
		return nil
	}
}

func EnforceContentLength(o *Options) error {
	o.EnforceContentLength = true
	return nil
}
