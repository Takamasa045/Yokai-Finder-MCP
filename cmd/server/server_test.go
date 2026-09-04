package main

import (
	"context"
	"strings"
	"testing"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
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
