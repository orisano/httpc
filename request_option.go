package httpc

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RequestOptions struct {
	Body    io.Reader
	Header  http.Header
	Queries url.Values

	EnforceContentLength bool
}

type RequestOption func(*RequestOptions) error

func ComposeRequestOption(opts ...RequestOption) RequestOption {
	return func(o *RequestOptions) error {
		for _, opt := range opts {
			if err := opt(o); err != nil {
				return err
			}
		}
		return nil
	}
}

func ApplyRequestOption(o *RequestOptions, opts ...RequestOption) error {
	return ComposeRequestOption(opts...)(o)
}

func WithBody(body io.Reader) RequestOption {
	return func(o *RequestOptions) error {
		o.Body = body
		return nil
	}
}

func WithBinary(bin io.Reader) RequestOption {
	return func(o *RequestOptions) error {
		o.Header.Add("Content-Type", "application/octet-stream")
		o.Body = bin
		return nil
	}
}

func WithForm(params url.Values) RequestOption {
	return func(o *RequestOptions) error {
		o.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		o.Body = strings.NewReader(params.Encode())
		return nil
	}
}

func WithJSON(data interface{}) RequestOption {
	return func(o *RequestOptions) error {
		o.Header.Set("Content-Type", "application/json")
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(data); err != nil {
			return err
		}
		o.Body = &b
		return nil
	}
}

func WithXML(data interface{}) RequestOption {
	return func(o *RequestOptions) error {
		o.Header.Set("Content-Type", `application/xml; charset="UTF-8"`)
		var b bytes.Buffer
		b.WriteString(xml.Header)
		if err := xml.NewEncoder(&b).Encode(data); err != nil {
			return err
		}
		o.Body = &b
		return nil
	}
}

func AddHeaderField(name, value string) RequestOption {
	return func(o *RequestOptions) error {
		o.Header.Add(name, value)
		return nil
	}
}

func WithHeader(header http.Header) RequestOption {
	return func(o *RequestOptions) error {
		o.Header = header
		return nil
	}
}

func AddQuery(key, value string) RequestOption {
	return func(o *RequestOptions) error {
		o.Queries.Add(key, value)
		return nil
	}
}

func WithQueries(queries url.Values) RequestOption {
	return func(o *RequestOptions) error {
		o.Queries = queries
		return nil
	}
}

func EnforceContentLength(o *RequestOptions) error {
	o.EnforceContentLength = true
	return nil
}
