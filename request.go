package httpc

import (
	"context"
	"net/http"
)

func NewRequest(ctx context.Context, method, rawurl string, opts ...RequestOption) (*http.Request, error) {
	c, err := NewClient(rawurl, nil)
	if err != nil {
		return nil, err
	}
	return c.NewRequest(ctx, method, "", opts...)
}
