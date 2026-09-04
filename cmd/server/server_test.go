package main

import (
	"context"
	"strings"
	"testing"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
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

	catalog, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "yokai://catalog"})
	if err != nil {
		t.Fatalf("catalog resource: %v", err)
	}
	if len(catalog.Contents) == 0 || !strings.Contains(catalog.Contents[0].Text, "indexCount") {
		t.Fatalf("unexpected catalog payload %+v", catalog)
	}
}

func TestYokaiNameFromURI(t *testing.T) {
	name, err := yokaiNameFromURI("yokai://yokai/河童")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if name != "河童" {
		t.Fatalf("got %q", name)
	}
}
