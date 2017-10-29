package httpc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/pkg/errors"
)

type debugTransport struct {
	w         io.Writer
	transport http.RoundTripper
}

func (d *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Fprintln(d.w, "debug-transport: ======== request ==========")
	if b, err := httputil.DumpRequestOut(req, true); err != nil {
		fmt.Fprintln(d.w, "debug-transport: failed to dump request:", err)
	} else {
		fmt.Fprintln(d.w, string(b))
	}

	resp, err := d.transport.RoundTrip(req)
	if err != nil {
		fmt.Fprintln(d.w, "debug-transport: failed to request:", err)
		return resp, err
	}

	fmt.Fprintln(d.w, "debug-transport: ======== response =========")
	if b, err := httputil.DumpResponse(resp, true); err != nil {
		fmt.Fprintln(d.w, "debug-transport: failed to dump response:", err)
	} else {
		fmt.Fprintln(d.w, string(b))
	}
	return resp, err
}

func InjectDebugTransport(client *http.Client, w io.Writer) error {
	if client == nil {
		return errors.New("missing client")
	}
	if w == nil {
		return errors.New("missing writer")
	}
	original := client.Transport
	if original == nil {
		original = http.DefaultTransport
	}
	client.Transport = &debugTransport{w: w, transport: original}
	return nil
}

func RemoveDebugTransport(client *http.Client) error {
	if client == nil {
		return errors.New("missing client")
	}
	if client.Transport == nil {
		return nil
	}
	if dt, ok := client.Transport.(*debugTransport); ok {
		client.Transport = dt.transport
	}
	return nil
}
