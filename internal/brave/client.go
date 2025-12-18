package brave

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const endpoint = "https://api.search.brave.com/res/v1/web/search"

type Client struct {
	hc *http.Client
}

func NewClient(hc *http.Client) *Client {
	return &Client{hc: hc}
}

func (c *Client) Search(ctx context.Context, apiKey string, req SearchRequest, debug bool) (SearchResponse, http.Header, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return SearchResponse{}, nil, err
	}
	q := u.Query()
	q.Set("q", req.Q)

	if req.Count > 0 {
		q.Set("count", fmt.Sprintf("%d", req.Count))
	}
	if req.Offset >= 0 {
		q.Set("offset", fmt.Sprintf("%d", req.Offset))
	}
	if req.SafeSearch != "" {
		q.Set("safesearch", req.SafeSearch)
	}
	if strings.TrimSpace(req.Freshness) != "" {
		q.Set("freshness", strings.TrimSpace(req.Freshness))
	}

	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return SearchResponse{}, nil, err
	}

	httpReq.Header.Set("Accept", "application/json")
	// IMPORTANT:
	// Do NOT set "Accept-Encoding" manually.
	// Go's net/http Transport will automatically request gzip and transparently decompress it
	// unless you override it. When you set it manually, you must decompress yourself.
	httpReq.Header.Set("X-Subscription-Token", apiKey)

	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return SearchResponse{}, nil, err
	}
	defer resp.Body.Close()

	headers := resp.Header.Clone()

	var bodyReader io.Reader = resp.Body

	// Safety net: if server still sends gzip, decode it.
	// (Some servers/proxies might enforce compression.)
	if ce := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Encoding"))); ce != "" {
		if strings.Contains(ce, "gzip") || strings.Contains(ce, "x-gzip") {
			gr, err := gzip.NewReader(resp.Body)
			if err != nil {
				return SearchResponse{}, headers, fmt.Errorf("failed to init gzip reader: %w", err)
			}
			defer gr.Close()
			bodyReader = gr
		}
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(bodyReader, 8192))
		msg := strings.TrimSpace(string(b))
		return SearchResponse{}, headers, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       msg,
		}
	}


	var out SearchResponse
	dec := json.NewDecoder(bodyReader)
	if err := dec.Decode(&out); err != nil {
		return SearchResponse{}, headers, fmt.Errorf("failed to decode json: %w", err)
	}

	return out, headers, nil
}
