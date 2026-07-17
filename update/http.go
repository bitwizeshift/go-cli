package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// fetchJSON performs a GET against url and decodes a successful JSON response
// into out. It returns [ErrUnexpectedStatus] for a non-2xx response and
// [ErrDecodeResponse] when the body cannot be decoded into out.
func fetchJSON(ctx context.Context, url string, out any) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, resp.Body.Close()) }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: %s", ErrUnexpectedStatus, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodeResponse, err)
	}
	return nil
}

// baseURL returns base without a trailing slash, or fallback when base is empty.
func baseURL(base, fallback string) string {
	if base == "" {
		return fallback
	}
	return strings.TrimSuffix(base, "/")
}
