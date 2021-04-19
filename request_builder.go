package httpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

type RequestBuilder struct {
	baseURL *url.URL
	header  http.Header
}

func NewRequestBuilder(rawurl string, header http.Header) (*RequestBuilder, error) {
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	return &RequestBuilder{
		baseURL: u,
		header:  header,
	}, nil
}

func (b *RequestBuilder) BaseURL() *url.URL {
	return b.baseURL
}

func (b *RequestBuilder) NewRequest(ctx context.Context, method, spath string, opts ...RequestOption) (*http.Request, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing ctx")
	}
	if method == "" {
		return nil, fmt.Errorf("missing method")
	}

	u := *b.baseURL
	if len(spath) > 0 {
		u.Path = path.Join(u.Path, spath)
	}

	h := cloneHeader(b.header)
	options := &RequestOptions{
		Header:  h,
		Queries: u.Query(),
	}

	if err := ApplyRequestOption(options, opts...); err != nil {
		return nil, fmt.Errorf("apply request option: %w", err)
	}

	u.RawQuery = options.Queries.Encode()

	contentLength := int64(0)
	if options.EnforceContentLength && options.Body != nil {
		type sizer interface {
			Size() int64
		}
		if s, ok := options.Body.(sizer); ok {
			contentLength = s.Size()
		} else {
			var b bytes.Buffer
			n, err := io.Copy(&b, options.Body)
			if err != nil {
				return nil, fmt.Errorf("read request body: %w", err)
			}
			contentLength = n
			options.Body = &b
		}
	}
	req, err := http.NewRequest(method, u.String(), options.Body)
	if err != nil {
		return nil, err
	}
	if options.EnforceContentLength {
		req.ContentLength = contentLength
	}
	req.Header = options.Header
	req = req.WithContext(ctx)

	return req, nil
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}
