package httpc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
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

	rt := d.transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	resp, err := rt.RoundTrip(req)
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
		return fmt.Errorf("missing client")
	}
	if w == nil {
		return fmt.Errorf("missing writer")
	}
	if t := client.Transport; t != nil {
		if _, ok := t.(*debugTransport); ok {
			return nil
		}
	}
	client.Transport = &debugTransport{w: w, transport: client.Transport}
	return nil
}

func RemoveDebugTransport(client *http.Client) error {
	if client == nil {
		return fmt.Errorf("missing client")
	}
	if client.Transport == nil {
		return nil
	}
	if dt, ok := client.Transport.(*debugTransport); ok {
		client.Transport = dt.transport
	}
	return nil
}
