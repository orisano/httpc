package httpc

import (
	"context"
	"net/http"
)

func NewRequest(ctx context.Context, method, rawurl string, opts ...RequestOption) (*http.Request, error) {
	b, err := NewRequestBuilder(rawurl, nil)
	if err != nil {
		return nil, err
	}
	return b.NewRequest(ctx, method, "", opts...)
}
