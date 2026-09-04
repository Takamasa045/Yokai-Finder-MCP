package ndl

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/version"
	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
	"golang.org/x/net/html"
)

const (
	ndlSearchURL = "https://ndlsearch.ndl.go.jp/api/opensearch"
	defaultLimit = 10
	maxLimit     = 100
	maxBodyBytes = 2 << 20
)

// Client wraps access to the NDL OpenSearch endpoint.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a client with sane defaults.
func NewClient() *Client {
	client := &http.Client{Timeout: 10 * time.Second}
	client.CheckRedirect = sameHostRedirect
	return &Client{
		baseURL:    ndlSearchURL,
		httpClient: client,
	}
}

func sameHostRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return fmt.Errorf("stopped after 5 redirects")
	}
	if len(via) == 0 {
		return nil
	}
	origin := via[0].URL
	if req.URL.Host != origin.Host || req.URL.Scheme != origin.Scheme {
		return fmt.Errorf("refusing redirect from %s to %s", origin.Host, req.URL.Host)
	}
	return nil
}

// WithHTTPClient allows injecting a custom http.Client (useful for testing).
func (c *Client) WithHTTPClient(client *http.Client) *Client {
	if client != nil {
		c.httpClient = client
	}
	return c
}

// WithBaseURL overrides the base API URL (useful for tests).
func (c *Client) WithBaseURL(baseURL string) *Client {
	if strings.TrimSpace(baseURL) != "" {
		c.baseURL = baseURL
	}
	return c
}

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	TotalResults string    `xml:"http://a9.com/-/spec/opensearchrss/1.0/ totalResults"`
	Items        []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string          `xml:"title"`
	Link        string          `xml:"link"`
	Description string          `xml:"description"`
	Author      string          `xml:"author"`
	DCTitle     string          `xml:"http://purl.org/dc/elements/1.1/ title"`
	Creator     string          `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Publisher   string          `xml:"http://purl.org/dc/elements/1.1/ publisher"`
	Date        string          `xml:"http://purl.org/dc/elements/1.1/ date"`
	Issued      string          `xml:"http://purl.org/dc/terms/ issued"`
	Subjects    []string        `xml:"http://purl.org/dc/elements/1.1/ subject"`
	Identifiers []xmlIdentifier `xml:"http://purl.org/dc/elements/1.1/ identifier"`
}

type xmlIdentifier struct {
	Type  string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Value string `xml:",chardata"`
}

// SearchYokaiBooks performs a search using the provided parameters.
func (c *Client) SearchYokaiBooks(ctx context.Context, params types.YokaiSearchParams) (*types.YokaiSearchResult, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("http client is not initialized")
	}

	query := buildQuery(params)

	limit := params.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	apiURL := c.buildAPIURL(query, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ndl api status %d", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxBodyBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if len(body) > maxBodyBytes {
		return nil, fmt.Errorf("ndl response exceeded %d bytes", maxBodyBytes)
	}

	result, err := parseRSS(body)
	if err != nil {
		return nil, err
	}

	result.Query = query
	return result, nil
}

func (c *Client) buildAPIURL(query string, limit int) string {
	params := url.Values{}
	params.Set("title", query)
	params.Set("cnt", fmt.Sprintf("%d", limit))
	return fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
}

func userAgent() string {
	return fmt.Sprintf("yokai-finder-mcp/%s (+https://github.com/Takamasa045/Yokai-Finder-MCP)", version.Version)
}

func buildQuery(params types.YokaiSearchParams) string {
	var parts []string
	if strings.TrimSpace(params.Name) != "" {
		parts = append(parts, strings.TrimSpace(params.Name))
	}
	if strings.TrimSpace(params.Region) != "" {
		parts = append(parts, strings.TrimSpace(params.Region))
	}
	if strings.TrimSpace(params.Category) != "" {
		parts = append(parts, strings.TrimSpace(params.Category))
	}
	if len(parts) == 0 {
		return "妖怪"
	}
	return strings.Join(parts, " ")
}

func parseRSS(data []byte) (*types.YokaiSearchResult, error) {
	var feed rssFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("parse rss: %w", err)
	}

	total := 0
	if trimmed := strings.TrimSpace(feed.Channel.TotalResults); trimmed != "" {
		if value, err := strconv.Atoi(trimmed); err == nil {
			total = value
		}
	}
	if total == 0 {
		total = len(feed.Channel.Items)
	}

	result := &types.YokaiSearchResult{
		Total:   total,
		Results: make([]types.YokaiBook, 0, len(feed.Channel.Items)),
	}

	for _, item := range feed.Channel.Items {
		book := types.YokaiBook{
			Title:       firstNonEmpty(item.Title, item.DCTitle),
			Author:      firstNonEmpty(item.Creator, item.Author),
			Publisher:   item.Publisher,
			PublishDate: firstNonEmpty(item.Issued, item.Date),
			Description: sanitizeDescription(item.Description),
			URL:         item.Link,
			Subjects:    item.Subjects,
			ISBN:        extractISBN(item.Identifiers),
		}
		book.CoverImageCandidates = BuildCoverURLs(book.ISBN, "").Candidates
		result.Results = append(result.Results, book)
	}

	return result, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func extractISBN(ids []xmlIdentifier) string {
	for _, id := range ids {
		if strings.Contains(strings.ToUpper(id.Type), "ISBN") {
			return strings.TrimSpace(id.Value)
		}
	}
	return ""
}

func sanitizeDescription(description string) string {
	if strings.TrimSpace(description) == "" {
		return ""
	}

	tokenizer := html.NewTokenizer(strings.NewReader(description))
	var builder strings.Builder

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return strings.TrimSpace(builder.String())
			}
			return strings.TrimSpace(description)
		case html.TextToken:
			builder.WriteString(tokenizer.Token().Data)
		}
	}
}
