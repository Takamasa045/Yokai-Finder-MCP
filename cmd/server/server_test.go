package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/ndl"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/yokai"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerListsExpectedToolsAndPrompts(t *testing.T) {
	ctx := context.Background()
	server := newServer(handler.New(nil, nil))

	t1, t2 := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	wantTools := map[string]bool{
		"suggest_yokai": true, "get_yokai": true, "list_yokai": true,
		"search_yokai_books": true, "yokai_of_the_day": true,
		"related_yokai": true, "compare_yokai": true,
		"list_curated_yokai": true, "get_cover_thumbnail": true,
	}
	got := map[string]bool{}
	for _, tool := range tools.Tools {
		got[tool.Name] = true
	}
	for name := range wantTools {
		if !got[name] {
			t.Errorf("missing tool %s", name)
		}
	}

	prompts, err := session.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}
	if len(prompts.Prompts) < 2 {
		t.Fatalf("expected story and character prompts, got %d", len(prompts.Prompts))
	}

	gotKappa, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_yokai",
		Arguments: map[string]any{"name": "カッパ"},
	})
	if err != nil {
		t.Fatalf("get_yokai: %v", err)
	}
	if gotKappa.IsError {
		t.Fatalf("get_yokai returned error: %+v", gotKappa)
	}

	completed, err := session.Complete(ctx, &mcp.CompleteParams{
		Ref:      &mcp.CompleteReference{Type: "ref/prompt", Name: "yokai_story"},
		Argument: mcp.CompleteParamsArgument{Name: "name", Value: "河"},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	found := false
	for _, v := range completed.Completion.Values {
		if v == "河童" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 河童 in completions, got %v", completed.Completion.Values)
	}

	uriComplete, err := session.Complete(ctx, &mcp.CompleteParams{
		Ref:      &mcp.CompleteReference{Type: "ref/resource", URI: "yokai://yokai/{name}"},
		Argument: mcp.CompleteParamsArgument{Name: "name", Value: "yokai://yokai/河"},
	})
	if err != nil {
		t.Fatalf("URI Complete: %v", err)
	}
	foundURI := false
	for _, v := range uriComplete.Completion.Values {
		if v == "yokai://yokai/河童" {
			foundURI = true
		}
	}
	if !foundURI {
		t.Fatalf("expected yokai://yokai/河童, got %v", uriComplete.Completion.Values)
	}

	catalog, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "yokai://catalog"})
	if err != nil {
		t.Fatalf("catalog resource: %v", err)
	}
	if len(catalog.Contents) == 0 || !strings.Contains(catalog.Contents[0].Text, "indexCount") {
		t.Fatalf("unexpected catalog payload %+v", catalog)
	}

	card, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "yokai://yokai/河童"})
	if err != nil {
		t.Fatalf("yokai card resource: %v", err)
	}
	if len(card.Contents) == 0 || !strings.Contains(card.Contents[0].Text, "河童") {
		t.Fatalf("unexpected yokai card %+v", card)
	}
	asciiCard, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "yokai://yokai/Kappa"})
	if err != nil {
		t.Fatalf("ASCII yokai card: %v", err)
	}
	if len(asciiCard.Contents) == 0 || !strings.Contains(asciiCard.Contents[0].Text, "河童") {
		t.Fatalf("unexpected Kappa card %+v", asciiCard)
	}

	prompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      "yokai_character_sheet",
		Arguments: map[string]string{"name": ""},
	})
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	if len(prompt.Messages) == 0 {
		t.Fatal("expected prompt message")
	}
	if text, ok := prompt.Messages[0].Content.(*mcp.TextContent); !ok || !strings.Contains(text.Text, "河童") {
		t.Fatalf("empty name should default to 河童, got %+v", prompt.Messages[0].Content)
	}

	listed, err := session.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources: %v", err)
	}
	wantURIs := 1 + len(yokai.Index())
	if got := len(listed.Resources); got != wantURIs {
		t.Fatalf("ListResources count %d want %d (catalog + index cards)", got, wantURIs)
	}
	seenCatalog, seenKappa := false, false
	for _, r := range listed.Resources {
		if r.URI == "yokai://catalog" {
			seenCatalog = true
		}
		if r.URI == "yokai://yokai/河童" {
			seenKappa = true
		}
	}
	if !seenCatalog || !seenKappa {
		t.Fatalf("ListResources missing catalog or 河童 card")
	}
}

func TestYokaiNameFromURI(t *testing.T) {
	tests := []struct {
		uri  string
		want string
	}{
		{"yokai://yokai/河童", "河童"},
		{"yokai://Kappa", "Kappa"},
	}
	for _, tc := range tests {
		name, err := yokaiNameFromURI(tc.uri)
		if err != nil {
			t.Fatalf("parse %s: %v", tc.uri, err)
		}
		if name != tc.want {
			t.Fatalf("%s: got %q want %q", tc.uri, name, tc.want)
		}
	}
	if _, err := yokaiNameFromURI("http://example/河童"); err == nil {
		t.Fatal("expected error for non-yokai scheme")
	}
}

func toolTestSession(t *testing.T, h *handler.Handler) (context.Context, *mcp.ClientSession) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	t1, t2 := mcp.NewInMemoryTransports()
	serverSession, err := newServer(h).Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return ctx, session
}

func toolPayload(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	if result.IsError {
		t.Fatalf("unexpected tool error: %+v", result)
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	// Older clients consume text content; it must carry the same JSON.
	if len(result.Content) != 1 {
		t.Fatalf("expected one JSON text block, got %+v", result.Content)
	}
	content, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !json.Valid([]byte(content.Text)) {
		t.Fatalf("invalid JSON text content: %+v", result.Content)
	}
	var textPayload map[string]any
	if err := json.Unmarshal([]byte(content.Text), &textPayload); err != nil || !reflect.DeepEqual(payload, textPayload) {
		t.Fatalf("text and structured content differ: %v", err)
	}
	return payload
}

func TestToolEmptyResultsAreArrays(t *testing.T) {
	ctx, session := toolTestSession(t, handler.New(nil, nil))
	for _, tc := range []struct {
		name string
		args map[string]any
		key  string
	}{
		{"list_yokai", map[string]any{"term": "zzzz-nope"}, "items"},
		{"list_yokai", map[string]any{"term": "---"}, "items"},
		{"list_yokai", map[string]any{"offset": 9999}, "items"},
		{"list_curated_yokai", map[string]any{"term": "zzzz-nope"}, "profiles"},
		{"related_yokai", map[string]any{"name": "zzzz-nope"}, "items"},
	} {
		t.Run(tc.name+"/"+tc.key+fmtArgs(tc.args), func(t *testing.T) {
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tc.name, Arguments: tc.args})
			if err != nil {
				t.Fatal(err)
			}
			payload := toolPayload(t, result)
			if items, ok := payload[tc.key].([]any); !ok || len(items) != 0 {
				t.Fatalf("%s should be an empty array, got %T", tc.key, payload[tc.key])
			}
		})
	}
}

func fmtArgs(args map[string]any) string {
	raw, _ := json.Marshal(args)
	return string(raw)
}

func TestToolInputErrorsAndRecovery(t *testing.T) {
	ctx, session := toolTestSession(t, handler.New(nil, nil))
	for _, tc := range []struct {
		name string
		args map[string]any
	}{
		{"get_yokai", map[string]any{}},
		{"get_yokai", map[string]any{"name": 123}},
		{"get_yokai", map[string]any{"name": "  "}},
		{"compare_yokai", map[string]any{"left": "河童", "right": " "}},
		{"get_cover_thumbnail", map[string]any{}},
		{"get_cover_thumbnail", map[string]any{"isbn": "invalid"}},
		{"get_cover_thumbnail", map[string]any{"jpno": "not-a-number"}},
		{"get_cover_thumbnail", map[string]any{"jpno": "123"}},
	} {
		t.Run(tc.name+fmtArgs(tc.args), func(t *testing.T) {
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tc.name, Arguments: tc.args})
			if err == nil && !result.IsError {
				t.Fatalf("invalid arguments must produce a protocol or tool error, got %+v", result)
			}
		})
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "get_yokai", Arguments: map[string]any{"name": "カッパ"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := toolPayload(t, result)["found"]; got != true {
		t.Fatalf("session must remain usable after errors: %+v", result)
	}
	for _, args := range []map[string]any{{"jpno": "22222222"}, {"isbn": "4-04-883992-6"}} {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "get_cover_thumbnail", Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		if len(toolPayload(t, result)["candidates"].([]any)) == 0 {
			t.Fatal("valid identifier must produce cover candidates")
		}
	}
}

type testRoundTripper func(*http.Request) (*http.Response, error)

func (f testRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNDLUnexpectedResponseIsNotCachedAsNoBooks(t *testing.T) {
	// A deterministic in-process transport: no sockets or external API calls.
	body := "<html><body>Maintenance</body></html>"
	calls := 0
	client := ndl.NewClient().WithHTTPClient(&http.Client{Transport: testRoundTripper(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})})
	c := cache.NewCache(time.Minute, 4)
	t.Cleanup(c.Stop)
	ctx, session := toolTestSession(t, handler.New(client, c))
	params := &mcp.CallToolParams{Name: "search_yokai_books", Arguments: map[string]any{"name": "河童"}}
	result, err := session.CallTool(ctx, params)
	if err != nil || !result.IsError {
		t.Fatalf("malformed NDL response should be a tool error: %+v %v", result, err)
	}
	daily, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "yokai_of_the_day", Arguments: map[string]any{"name": "河童"}})
	if err != nil {
		t.Fatal(err)
	}
	payload := toolPayload(t, daily)
	if payload["profile"].(map[string]any)["nativeName"] != "河童" || len(payload["notes"].([]any)) == 0 {
		t.Fatalf("daily pick should preserve lore and explain NDL failure: %+v", payload)
	}
	body = "<rss><channel/></rss>"
	before := calls
	for i := 0; i < 2; i++ {
		result, err := session.CallTool(ctx, params)
		if err != nil {
			t.Fatal(err)
		}
		payload := toolPayload(t, result)
		if payload["total"] != float64(0) || len(payload["results"].([]any)) != 0 {
			t.Fatalf("valid zero-result RSS should succeed: %+v", payload)
		}
	}
	if calls != before+1 {
		t.Fatalf("error must not be cached; successful empty result should be cached, calls before=%d after=%d", before, calls)
	}
}
