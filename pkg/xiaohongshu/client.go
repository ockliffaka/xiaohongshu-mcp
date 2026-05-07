// Package xiaohongshu provides a client for interacting with the Xiaohongshu (Little Red Book) platform.
package xiaohongshu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is the base URL for the Xiaohongshu API.
	DefaultBaseURL = "https://www.xiaohongshu.com"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultUserAgent mimics a browser to avoid being blocked.
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// Client is the Xiaohongshu HTTP client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	cookies    []*http.Cookie
}

// Note represents a Xiaohongshu post/note.
type Note struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"desc"`
	Author      Author `json:"user"`
	LikeCount   int    `json:"liked_count"`
	CollectCount int   `json:"collected_count"`
	CommentCount int   `json:"comment_count"`
	Images      []Image `json:"image_list"`
	Type        string `json:"type"`
	CreateTime  int64  `json:"time"`
}

// Author represents the author of a note.
type Author struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// Image represents an image in a note.
type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SearchResult represents the result of a search query.
type SearchResult struct {
	Notes []*Note `json:"notes"`
	Total int     `json:"total"`
	Page  int     `json:"page"`
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the client.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithCookies sets cookies for authenticated requests.
func WithCookies(cookies []*http.Cookie) ClientOption {
	return func(c *Client) {
		c.cookies = cookies
	}
}

// WithTimeout sets a custom timeout for the HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Xiaohongshu client with the given options.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SearchNotes searches for notes by keyword.
func (c *Client) SearchNotes(ctx context.Context, keyword string, page int) (*SearchResult, error) {
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", "20")

	reqURL := fmt.Sprintf("%s/api/sns/web/v1/search/notes?%s", c.baseURL, params.Encode())

	resp, err := c.doRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("search notes request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search result: %w", err)
	}

	return &result, nil
}

// doRequest performs an HTTP request with the client's default headers and cookies.
func (c *Client) doRequest(ctx context.Context, method, reqURL string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", c.baseURL)
	req.Header.Set("Origin", c.baseURL)

	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}
