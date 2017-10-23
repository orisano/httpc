package httpc

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func NewRequest(ctx context.Context, method, rawurl string, opts ...Option) (*http.Request, error) {
	if ctx == nil {
		return nil, errors.New("missing ctx")
	}
	if len(method) == 0 {
		return nil, errors.New("missing method")
	}
	if len(rawurl) == 0 {
		return nil, errors.New("missing rawurl")
	}

	options := &Options{
		Header: make(http.Header),
		Params: make(url.Values),
	}
	if err := ApplyOption(options, opts...); err != nil {
		return nil, errors.Wrap(err, "failed to apply option")
	}

	if len(options.Params) > 0 {
		u, err := url.ParseRequestURI(rawurl)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse url: %s", rawurl)
		}
		q := u.Query()
		addValues(q, options.Params)
		u.RawQuery = q.Encode()
		rawurl = u.String()
	}

	contentLength := int64(0)
	if options.EnforceContentLength && options.Body != nil {
		type sizer interface {
			Size() int64
		}
		if s, ok := options.Body.(sizer); ok {
			contentLength = s.Size()
		} else {
			b, err := ioutil.ReadAll(options.Body)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read body")
			}
			contentLength = int64(len(b))
			options.Body = bytes.NewReader(b)
		}
	}
	req, err := http.NewRequest(method, rawurl, options.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to http request construct")
	}
	if options.EnforceContentLength {
		req.ContentLength = contentLength
	}
	req.Header = options.Header
	req = req.WithContext(ctx)

	return req, nil
}

func addValues(dst, src url.Values) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}
