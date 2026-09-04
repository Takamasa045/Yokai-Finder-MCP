package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerResources(server *mcp.Server, h *handler.Handler) {
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "yokai-card",
		Title:       "Yokai card",
		MIMEType:    "application/json",
		URITemplate: "yokai://yokai/{name}",
		Description: "Encyclopedia or index card for one yokai. Example: yokai://yokai/河童",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		name, err := yokaiNameFromURI(req.Params.URI)
		if err != nil {
			return nil, err
		}
		result, err := h.GetYokai(ctx, types.GetYokaiParams{Name: name})
		if err != nil {
			return nil, err
		}
		if !result.Found {
			return nil, mcp.ResourceNotFoundError(req.Params.URI)
		}
		payload, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(payload),
			}},
		}, nil
	})
}

func registerPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "yokai_story",
		Title:       "Yokai short story",
		Description: "Write a short scene featuring a named yokai, using get_yokai lore first.",
		Arguments: []*mcp.PromptArgument{{
			Name:        "name",
			Description: "Yokai name (Japanese or English)",
			Required:    true,
		}},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name := strings.TrimSpace(req.Params.Arguments["name"])
		if name == "" {
			name = "河童"
		}
		text := fmt.Sprintf("妖怪ファインダーの get_yokai で「%s」を調べてから、伝承を歪めすぎない短い場面（600字以内）を書いてください。出典が図鑑にない想像は創作だと明記してください。", name)
		return &mcp.GetPromptResult{
			Description: "Short folklore-aware scene",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: text},
			}},
		}, nil
	})

	server.AddPrompt(&mcp.Prompt{
		Name:        "yokai_character_sheet",
		Title:       "Yokai character sheet",
		Description: "Turn a yokai into a fiction character sheet without erasing the folklore.",
		Arguments: []*mcp.PromptArgument{{
			Name:        "name",
			Description: "Yokai name",
			Required:    true,
		}},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name := strings.TrimSpace(req.Params.Arguments["name"])
		text := fmt.Sprintf("get_yokai で「%s」を取得し、外見・動機・弱点・現代での居場所・やってはいけないこと、の5項目でキャラクターシートを作ってください。図鑑にない点は推測と書いてください。", name)
		return &mcp.GetPromptResult{
			Description: "Character sheet from folklore",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: text},
			}},
		}, nil
	})
}

func yokaiNameFromURI(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme != "yokai" {
		return "", mcp.ResourceNotFoundError(raw)
	}
	name := strings.TrimPrefix(u.Path, "/")
	if u.Host == "yokai" && name != "" {
		return name, nil
	}
	if u.Host != "" && u.Host != "yokai" {
		return u.Host, nil
	}
	if name != "" {
		return name, nil
	}
	if u.Opaque != "" {
		return strings.TrimPrefix(u.Opaque, "yokai/"), nil
	}
	return "", mcp.ResourceNotFoundError(raw)
}
