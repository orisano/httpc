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
	Body    io.Reader
	Header  http.Header
	Queries url.Values

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

func AddHeaderField(name, value string) Option {
	return func(o *Options) error {
		o.Header.Add(name, value)
		return nil
	}
}

func WithHeader(header http.Header) Option {
	return func(o *Options) error {
		o.Header = header
		return nil
	}
}

func AddQuery(key, value string) Option {
	return func(o *Options) error {
		o.Queries.Add(key, value)
		return nil
	}
}

func WithQueries(queries url.Values) Option {
	return func(o *Options) error {
		o.Queries = queries
		return nil
	}
}

func EnforceContentLength(o *Options) error {
	o.EnforceContentLength = true
	return nil
}
