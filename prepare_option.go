package httpc

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type PrepareOptions struct {
	Body    io.Reader
	Header  http.Header
	Queries url.Values

	EnforceContentLength bool
}

type PrepareOption func(*PrepareOptions) error

func ComposePrepareOption(opts ...PrepareOption) PrepareOption {
	return func(o *PrepareOptions) error {
		for _, opt := range opts {
			if err := opt(o); err != nil {
				return err
			}
		}
		return nil
	}
}

func ApplyPrepareOption(o *PrepareOptions, opts ...PrepareOption) error {
	return ComposePrepareOption(opts...)(o)
}

func WithBody(body io.Reader) PrepareOption {
	return func(o *PrepareOptions) error {
		o.Body = body
		return nil
	}
}

func WithBinary(bin io.Reader) PrepareOption {
	return func(o *PrepareOptions) error {
		o.Header.Add("Content-Type", "application/octet-stream")
		o.Body = bin
		return nil
	}
}

func WithForm(params url.Values) PrepareOption {
	return func(o *PrepareOptions) error {
		o.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		o.Body = strings.NewReader(params.Encode())
		return nil
	}
}

func WithJSON(data interface{}) PrepareOption {
	return func(o *PrepareOptions) error {
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

func WithXML(data interface{}) PrepareOption {
	return func(o *PrepareOptions) error {
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

func AddHeaderField(name, value string) PrepareOption {
	return func(o *PrepareOptions) error {
		o.Header.Add(name, value)
		return nil
	}
}

func WithHeader(header http.Header) PrepareOption {
	return func(o *PrepareOptions) error {
		o.Header = header
		return nil
	}
}

func AddQuery(key, value string) PrepareOption {
	return func(o *PrepareOptions) error {
		o.Queries.Add(key, value)
		return nil
	}
}

func WithQueries(queries url.Values) PrepareOption {
	return func(o *PrepareOptions) error {
		o.Queries = queries
		return nil
	}
}

func EnforceContentLength(o *PrepareOptions) error {
	o.EnforceContentLength = true
	return nil
}
